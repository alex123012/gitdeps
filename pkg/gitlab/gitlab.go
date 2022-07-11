package gitlab

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alex123012/gitdeps/cmd/common"
	"github.com/alex123012/gitdeps/pkg/client"
	"github.com/xanzy/go-gitlab"
)

func GetHostByAnnotation(gl *client.Client, hostUrl string) (*client.Host, error) {

	host, f := gl.Hosts.GetHost(hostUrl)
	if !f {
		return nil, fmt.Errorf("no host found from annotation")
	}
	return host, nil
}

func GetBranchByPipeline(host *client.Host, projectPath string, pipelineNumber int) (string, error) {
	pipeline, _, err := host.Client.Pipelines.GetPipeline(projectPath, pipelineNumber)
	if err != nil {
		return "", err
	}

	return pipeline.Ref, nil
}

func GetDefaultBranch(host *client.Host, projectPath string) (string, error) {
	project, _, err := host.Client.Projects.GetProject(projectPath, &gitlab.GetProjectOptions{})
	if err != nil {
		return "", err
	}

	return project.DefaultBranch, nil
}

func CompareBranches(host *client.Host, projectPath, defaultBranch, targetBranch string) (*gitlab.Compare, error) {
	straight := false // TODO Make flag
	opts := &gitlab.CompareOptions{
		From:     &targetBranch,
		To:       &defaultBranch,
		Straight: &straight,
	}
	compare, _, err := host.Client.Repositories.Compare(projectPath, opts)
	if err != nil {
		return nil, err
	}
	return compare, nil
}
func TargetHaveAllCommitsFromDefault(annotationValue string) (bool, error) {
	splitUrl := strings.Split(client.TrimUrl(annotationValue), "/")

	hostUrl := splitUrl[0]
	host, err := GetHostByAnnotation(common.Client, hostUrl)
	if err != nil {
		return false, err
	}

	pipelineNumber, err := strconv.Atoi(splitUrl[len(splitUrl)-1])
	if err != nil {
		return false, err
	}
	projectPath := strings.Join(splitUrl[1:len(splitUrl)-2], "/")
	targetBranch, err := GetBranchByPipeline(host, projectPath, pipelineNumber)
	if err != nil {
		return false, err
	}

	defaultBranch, err := GetDefaultBranch(host, projectPath)
	if err != nil {
		return false, err
	}

	if defaultBranch == targetBranch {
		return true, nil
	}

	compare, err := CompareBranches(host, projectPath, defaultBranch, targetBranch)

	if err != nil {
		return false, err
	}

	if len(compare.Diffs) > 0 {
		return false, nil
	}
	return true, nil
}
