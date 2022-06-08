package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"k8s.io/klog/v2"
)

var (
	GET  = "GET"
	POST = "POST"
)

type GitlabPageInterfaces interface {
	GitlabRepository | GitlabBranch | GitlabMergeRequest
}

type GitlabController struct {
	name               string
	Repositories       []GitlabRepository
	DefaultTimer       time.Duration
	DefaultTimerRepo   time.Duration
	DefaultSleepApi    time.Duration
	RetriesOnFailure   int
	token              string
	tokenHeader        string
	defaultApiUri      string
	url                string
	oldRepositoriesMap map[*GitlabRepository]bool
	repositoriesMap    map[*GitlabRepository]bool
	repoMux            *sync.Mutex
	requestMux         *sync.Mutex
	repoMapMux         *sync.RWMutex
	client             *http.Client
	errorGroup         *errgroup.Group
}

type GitlabRepository struct {
	Name          string `json:"name"`
	Id            int    `json:"id"`
	DefaultBranch string `json:"default_branch"`
	Branches      []GitlabBranch
	MergeRequests map[string]bool
	Controller    *GitlabController
}

type GitlabCompare struct {
	Diffs []GitlabDiff `json:"diffs"`
}

type GitlabMergeRequest struct {
	TargetBranch string `json:"target_branch"`
	SourceBranch string `json:"source_branch"`
}

type GitlabBranch struct {
	Name string `json:"name"`
}

type GitlabDiff struct {
}

func (g *GitlabController) insertInRepoMap(key *GitlabRepository, value bool) {
	g.repoMapMux.Lock()
	defer g.repoMapMux.Unlock()
	g.repositoriesMap[key] = value
}
func (g *GitlabController) MakeApiRequest(method, url string, body io.Reader, jsonVar interface{}) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		klog.Fatalln(err)
	}
	req.Header.Set(g.tokenHeader, g.token)
	retries := g.RetriesOnFailure
	for i := 0; i < retries; i++ {
		g.requestMux.Lock()
		resp, err := g.client.Do(req)
		time.Sleep(g.DefaultSleepApi)
		g.requestMux.Unlock()

		if err != nil {
			if i == (retries - 1) {
				klog.Fatalln(err)
			} else {
				klog.Errorln(err)
				continue
			}
		}

		if resp.StatusCode > 299 {
			klog.Errorf("Response failed with status code: %s on url %s", resp.Status, url)

			continue
		}
		defer resp.Body.Close()
		return json.NewDecoder(resp.Body).Decode(jsonVar)
	}
	return fmt.Errorf("error getting response from %s", url)

}

func getAllPerPage[V GitlabPageInterfaces](g *GitlabController, url string, variable []V) ([]V, error) {
	var i int
	for {
		var iter []V
		err := g.MakeApiRequest(GET, fmt.Sprintf(url, i), nil, &iter)
		if err != nil {
			return variable, jsonErrorCatch(err)
		}
		variable = append(variable, iter...)
		i += 1
		if len(iter) < 100 {
			break
		}
	}
	return variable, nil
}
func (g *GitlabController) GetAllProjects(ctx context.Context, waitChan chan<- error) error {
	defer close(waitChan)
	klog.Infoln("Updating repository list")
	url := g.url + "/projects?per_page=100&page=%d"
	var projects []GitlabRepository
	projects, err := getAllPerPage(g, url, projects)
	if err != nil {
		waitChan <- err
	}
	klog.Infof("There are %d projects in this GitLab", len(projects))
	g.repoMux.Lock()
	defer g.repoMux.Unlock()
	g.Repositories = projects
	waitChan <- nil
	return nil
}
func (g *GitlabController) UpdateAllProjects(ctx context.Context) error {
	g.repoMapMux.Lock()
	g.oldRepositoriesMap = make(map[*GitlabRepository]bool)
	for k, v := range g.repositoriesMap {
		g.oldRepositoriesMap[k] = v
	}
	g.repoMapMux.Unlock()

	waitChan := make(chan error)
	go g.GetAllProjects(ctx, waitChan)
	<-waitChan

	g.repoMapMux.Lock()
	i := 0
	keys := make([]*GitlabRepository, len(g.repositoriesMap))
	for k := range g.repositoriesMap {
		keys[i] = k
		i++
	}
	g.repoMapMux.Unlock()

	for _, key := range keys {
		if value, f := g.oldRepositoriesMap[key]; !f || !value {
			tmp := key
			klog.Infof("Starting following project %s", tmp.Name)
			g.errorGroup.Go(FuncDecorator(ctx, tmp.Run, g.DefaultTimer))
			g.insertInRepoMap(tmp, true)
		}
	}
	return nil
}
func (g *GitlabController) SetToken(token string) {
	g.token = token
}
func (g *GitlabController) SetUrl(url string) {
	https, http := "https://", "http://"
	url = strings.TrimRight(url, "/")
	url = strings.TrimRight(url, g.defaultApiUri)

	if strings.HasPrefix(url, http) {
		url = strings.Replace(url, http, https, 1)
	} else if !strings.HasPrefix(url, https) {
		url = https + url
	}
	g.url = url + g.defaultApiUri
	klog.Infof("GitLab api url: %s", g.url)
}

func (g *GitlabController) GetUrl() string {
	return g.url
}
func (g *GitlabController) Run(ctx context.Context) error {
	waitChan := make(chan error)
	go g.GetAllProjects(ctx, waitChan)
	select {
	case err := <-waitChan:
		if err != nil {
			klog.Warningf("Repo list created with error: %s", err)
		}
	case <-ctx.Done():
		klog.Warningln("Exiting...")
	}
	g.errorGroup, ctx = errgroup.WithContext(ctx)

	for _, repo := range g.Repositories {
		tmp := repo
		tmp.Controller = g
		g.errorGroup.Go(FuncDecorator(ctx, tmp.Run, g.DefaultTimer))
	}
	g.errorGroup.Go(FuncDecorator(ctx, g.UpdateAllProjects, g.DefaultTimerRepo))
	return g.errorGroup.Wait()
}

func (g *GitlabController) GetName() string {
	return g.name
}

func New(
	token string,
	url string,
	defaultTimer int,
	defaultTimerRepo int,
	defaultSleepApi int,
	retriesOnFailure int,
	repositoriesIds []int) *GitlabController {

	ctr := &GitlabController{
		name:             "Gitlab",
		tokenHeader:      "PRIVATE-TOKEN",
		defaultApiUri:    "/api/v4",
		repoMux:          &sync.Mutex{},
		requestMux:       &sync.Mutex{},
		repoMapMux:       &sync.RWMutex{},
		client:           &http.Client{},
		DefaultTimer:     time.Duration(defaultTimer) * time.Second,
		DefaultTimerRepo: time.Duration(defaultTimerRepo) * time.Second,
		DefaultSleepApi:  time.Duration(defaultSleepApi) * time.Millisecond,
		RetriesOnFailure: retriesOnFailure,
	}
	ctr.repositoriesMap = make(map[*GitlabRepository]bool)
	ctr.SetToken(token)
	ctr.SetUrl(url)
	return ctr
}

func (g *GitlabRepository) CompareBranches(source, target string) (bool, error) {
	url := fmt.Sprintf("%s/projects/%d/repository/compare?to=%s&from=%s", g.Controller.url, g.Id, source, target)
	var compare GitlabCompare
	err := g.Controller.MakeApiRequest(GET, url, nil, &compare)
	return len(compare.Diffs) > 0, err
}

func (g *GitlabRepository) GetAllMRs(ctx context.Context) error {
	klog.Infof("Updating %s repository MR list (id: %d)", g.Name, g.Id)
	url := g.Controller.url + "/projects/" + fmt.Sprint(g.Id) + "/merge_requests?per_page=100&page=%d"
	var mrs []GitlabMergeRequest
	mrs, err := getAllPerPage(g.Controller, url, mrs)
	if err != nil {
		return err
	}
	mrMap := make(map[string]bool)
	for _, iter := range mrs {
		mrMap[iter.SourceBranch+iter.TargetBranch] = true
	}
	g.MergeRequests = mrMap
	klog.Infof("There are %d MRs in project %s (id: %d)", len(mrs), g.Name, g.Id)
	return nil
}

func (g *GitlabRepository) GetAllBranches(ctx context.Context) error {

	klog.Infof("Updating %s repository branches list (id: %d)", g.Name, g.Id)
	url := g.Controller.url + "/projects/" + fmt.Sprint(g.Id) + "/repository/branches?per_page=100&page=%d"
	var branches []GitlabBranch
	branches, err := getAllPerPage(g.Controller, url, branches)
	if err != nil {
		return err
	}
	g.Branches = branches
	klog.Infof("There are %d branches in project %s (id: %d)", len(branches), g.Name, g.Id)
	return nil
}

func (g *GitlabRepository) hasThisMr(source, target string) bool {
	_, found := g.MergeRequests[source+target]
	return found
}

func (g *GitlabRepository) Run(ctx context.Context) error {
	err := g.GetAllBranches(ctx)
	if err != nil {
		g.Controller.insertInRepoMap(g, false)
		return err
	}
	err = g.GetAllMRs(ctx)
	if err != nil {
		g.Controller.insertInRepoMap(g, false)
		return err
	}
	klog.Infof("Checking branches count in project %s (id: %d)", g.Name, g.Id)
	if len(g.Branches) < 2 {
		g.Controller.insertInRepoMap(g, false)
		return nil
	}
	klog.Infof("Comparing branches in project %s (id: %d)", g.Name, g.Id)
	for _, branch := range g.Branches {
		if branch.Name == g.DefaultBranch {
			continue
		}
		res, err := g.CompareBranches(branch.Name, g.DefaultBranch)
		if err != nil {
			g.Controller.insertInRepoMap(g, false)
			return err
		}

		if !res && !g.hasThisMr(branch.Name, g.DefaultBranch) {
			klog.Infof("There is divergency in project %s (id: %d) between default branch (%s) and branch %s", g.Name, g.Id, g.DefaultBranch, branch.Name)
		}
	}
	return nil
}

func FuncDecorator(ctx context.Context, f func(ctx context.Context) error, timer time.Duration) func() error {
	return func() error {
		go TimerFunc(ctx, f, timer)
		<-ctx.Done()
		return nil
	}
}

func TimerFunc(ctx context.Context, f func(ctx context.Context) error, timer time.Duration) {
	for {
		err := f(ctx)
		if err != nil {
			klog.Errorln(err)
		}
		timer := time.NewTimer(timer)
		<-timer.C
		timer.Reset(time.Second * 0)
	}
}
func jsonErrorCatch(err error) error {
	klog.Errorf("error decoding gitlab response: %v", err)
	klog.Errorln("Check PRIVATE TOKEN validity")

	if e, ok := err.(*json.SyntaxError); ok {
		klog.Errorf("syntax error at byte offset %d", e.Offset)
	}
	return err
}
