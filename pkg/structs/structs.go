package structs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

type Gitlab interface {
	GitlabRepository | GitlabBranchJson | GitlabMergeRequestJson
}

type ControllerInterface interface {
	getJsonStruct(method, url string, body io.Reader, jsonVar interface{}) error
	GetAllProjects(ctx context.Context) error
	SetToken(token string)
	SetUrl(url string)
	Run(url string)
}

type RepositoryInterface interface {
	CompareBranches(source, feature string)
	GetAllMRs(ctx context.Context)
	GetAllBranches(ctx context.Context) error
	CompareInfo()
	Run()
}

type GitlabController struct {
	token              string
	headerToken        string
	url                string
	defaultApiUri      string
	defaultTimer       time.Duration
	defaultTimerRepo   time.Duration
	defaultSleepApi    time.Duration
	oldRepositoriesMap map[*GitlabRepository]bool
	RepositoriesMap    map[*GitlabRepository]bool
	Repositories       []GitlabRepository
	repoMux            *sync.Mutex
	requestMux         *sync.Mutex
	client             *http.Client
	errorGroup         *errgroup.Group
}

func (g *GitlabController) getJsonStruct(method, url string, body io.Reader, jsonVar interface{}) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set(g.headerToken, g.token)

	g.requestMux.Lock()
	resp, err := g.client.Do(req)
	time.Sleep(g.defaultSleepApi)
	g.requestMux.Unlock()
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(jsonVar)
}
func getAllPerPage[V Gitlab](g *GitlabController, url string, variable []V) ([]V, error) {
	var i int
	for {
		var iter []V
		err := g.getJsonStruct(GET, fmt.Sprintf(url, i), nil, &iter)
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
func (g *GitlabController) GetAllProjects(ctx context.Context) error {
	klog.Infoln("Updating repository list")
	url := g.url + "/projects?per_page=100&page=%d"
	var projects []GitlabRepository
	projects, err := getAllPerPage(g, url, projects)
	if err != nil {
		return err
	}
	klog.Infof("There are %d projects in this GitLab", len(projects))
	g.repoMux.Lock()
	defer g.repoMux.Unlock()
	g.Repositories = projects
	return nil
}
func (g *GitlabController) updateAllProjects(ctx context.Context) error {
	g.oldRepositoriesMap = g.RepositoriesMap
	g.GetAllProjects(ctx)
	for key := range g.RepositoriesMap {
		if value, f := g.oldRepositoriesMap[key]; f && !value {
			tmp := key
			g.errorGroup.Go(timerFuncDecorator(ctx, tmp.Run, g.defaultTimer))
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
	fmt.Println(url)
	g.url = url + g.defaultApiUri
	fmt.Println(g.url)
}
func (g *GitlabController) Run(ctx context.Context) error {
	err := g.GetAllProjects(ctx)
	if err != nil {
		return err
	}
	g.errorGroup, ctx = errgroup.WithContext(ctx)
	//  = gr
	for _, repo := range g.Repositories {
		tmp := repo
		tmp.Controller = g
		g.errorGroup.Go(timerFuncDecorator(ctx, tmp.Run, g.defaultTimer))
	}
	g.errorGroup.Go(timerFuncDecorator(ctx, g.updateAllProjects, g.defaultTimerRepo))
	return g.errorGroup.Wait()
}

func NewGitLab(token, url string) *GitlabController {

	ctr := &GitlabController{
		headerToken:      "PRIVATE-TOKEN",
		defaultApiUri:    "/api/v4",
		repoMux:          &sync.Mutex{},
		requestMux:       &sync.Mutex{},
		client:           &http.Client{},
		defaultTimer:     time.Duration(10) * time.Second,
		defaultTimerRepo: time.Duration(40) * time.Second,
		defaultSleepApi:  time.Duration(300) * time.Millisecond,
	}
	ctr.RepositoriesMap = make(map[*GitlabRepository]bool)
	ctr.SetToken(token)
	ctr.SetUrl(url)
	return ctr
}

type GitlabRepository struct {
	Name          string `json:"name"`
	Id            int    `json:"id"`
	DefaultBranch string `json:"default_branch"`
	Branches      []GitlabBranchJson
	MergeRequests map[string]bool
	Controller    *GitlabController
}

func (g *GitlabRepository) CompareBranches(source, target string) (bool, error) {
	url := fmt.Sprintf("%s/projects/%d/repository/compare?to=%s&from=%s", g.Controller.url, g.Id, source, target)
	var compare GitlabCompareJson
	err := g.Controller.getJsonStruct(GET, url, nil, &compare)
	return len(compare.Diffs) > 0, err
}
func (g *GitlabRepository) GetAllMRs(ctx context.Context) error {
	// klog.Infof("Updating %s repository MR list (id: %d)", g.Name, g.Id)
	url := g.Controller.url + "/projects/" + fmt.Sprint(g.Id) + "/merge_requests?per_page=100&page=%d"
	var mrs []GitlabMergeRequestJson
	mrs, err := getAllPerPage(g.Controller, url, mrs)
	if err != nil {
		return err
	}
	mrMap := make(map[string]bool)
	for _, iter := range mrs {
		mrMap[iter.SourceBranch+iter.TargetBranch] = true
	}
	g.MergeRequests = mrMap
	// klog.Infof("There are %d MRs in project %s (id: %d)", len(mrs), g.Name, g.Id)
	return nil
}
func (g *GitlabRepository) GetAllBranches(ctx context.Context) error {

	// klog.Infof("Updating %s repository branches list (id: %d)", g.Name, g.Id)
	url := g.Controller.url + "/projects/" + fmt.Sprint(g.Id) + "/repository/branches?per_page=100&page=%d"
	var branches []GitlabBranchJson
	branches, err := getAllPerPage(g.Controller, url, branches)
	if err != nil {
		return err
	}
	g.Branches = branches
	// klog.Infof("There are %d branches in project %s (id: %d)", len(branches), g.Name, g.Id)
	return nil
}
func (g *GitlabRepository) hasThisMr(source, target string) bool {
	_, found := g.MergeRequests[source+target]
	return found
}
func (g *GitlabRepository) Run(ctx context.Context) error {
	err := g.GetAllBranches(ctx)
	if err != nil {
		g.Controller.RepositoriesMap[g] = false
		return err
	}
	err = g.GetAllMRs(ctx)
	if err != nil {
		g.Controller.RepositoriesMap[g] = false
		return err
	}
	// klog.Infoln("Checking branches count in project", g.Name)
	if len(g.Branches) < 2 {
		g.Controller.RepositoriesMap[g] = false
		return nil
	}
	klog.Infoln("Comparing branches in project", g.Name)
	for _, branch := range g.Branches {
		if branch.Name == g.DefaultBranch {
			continue
		}
		res, err := g.CompareBranches(branch.Name, g.DefaultBranch)
		if err != nil {
			g.Controller.RepositoriesMap[g] = false
			return err
		}

		if !res && !g.hasThisMr(branch.Name, g.DefaultBranch) {
			klog.Infof("There is divergency in project %s between default branch (%s) and branch %s", g.Name, g.DefaultBranch, branch.Name)
		} else {
			klog.Infof("All good in project %s between default branch (%s) and branch %s", g.Name, g.DefaultBranch, branch.Name)
		}
	}
	return nil
}

func timerFuncDecorator(ctx context.Context, f func(ctx context.Context) error, timer time.Duration) func() error {
	return func() error {
		for {
			timer := time.NewTimer(timer)
			<-timer.C
			err := f(ctx)
			if err != nil {
				return err
			}
			timer.Reset(time.Second * 0)
		}
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
