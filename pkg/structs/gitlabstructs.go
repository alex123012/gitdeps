package structs

import "time"

type GitlabCompareJson struct {
	Commit         GitlabCommitJson   `json:"commit"`
	Commits        []GitlabCommitJson `json:"commits"`
	Diffs          []GitlabDiffJson   `json:"diffs"`
	CompareTimeout bool               `json:"compare_timeout"`
	CompareSameRef bool               `json:"compare_same_ref"`
	WebURL         string             `json:"web_url"`
}

type GitlabUserJson struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	State     string `json:"state"`
	AvatarURL string `json:"avatar_url"`
	WebURL    string `json:"web_url"`
}

type GitlabMergeRequestJson struct {
	ID                          int                            `json:"id"`
	Iid                         int                            `json:"iid"`
	ProjectID                   int                            `json:"project_id"`
	Title                       string                         `json:"title"`
	Description                 string                         `json:"description"`
	State                       string                         `json:"state"`
	CreatedAt                   time.Time                      `json:"created_at"`
	UpdatedAt                   time.Time                      `json:"updated_at"`
	MergedBy                    GitlabUserJson                 `json:"merged_by,omitempty"`
	MergeUser                   GitlabUserJson                 `json:"merge_user,omitempty"`
	MergedAt                    string                         `json:"merged_at,omitempty"`
	ClosedBy                    GitlabUserJson                 `json:"closed_by,omitempty"`
	ClosedAt                    string                         `json:"closed_at,omitempty"`
	TargetBranch                string                         `json:"target_branch"`
	SourceBranch                string                         `json:"source_branch"`
	UserNotesCount              int                            `json:"user_notes_count"`
	Upvotes                     int                            `json:"upvotes"`
	Downvotes                   int                            `json:"downvotes"`
	Author                      GitlabUserJson                 `json:"author"`
	Assignees                   []GitlabUserJson               `json:"assignees"`
	Assignee                    GitlabUserJson                 `json:"assignee"`
	Reviewers                   []GitlabUserJson               `json:"reviewers,omitempty"`
	SourceProjectID             int                            `json:"source_project_id"`
	TargetProjectID             int                            `json:"target_project_id"`
	Labels                      []string                       `json:"labels,omitempty"`
	Draft                       bool                           `json:"draft"`
	WorkInProgress              bool                           `json:"work_in_progress"`
	Milestone                   string                         `json:"milestone,omitempty"`
	MergeWhenPipelineSucceeds   bool                           `json:"merge_when_pipeline_succeeds"`
	MergeStatus                 string                         `json:"merge_status"`
	Sha                         string                         `json:"sha"`
	MergeCommitSha              string                         `json:"merge_commit_sha,omitempty"`
	SquashCommitSha             string                         `json:"squash_commit_sha,omitempty"`
	DiscussionLocked            string                         `json:"discussion_locked,omitempty"`
	ShouldRemoveSourceBranch    bool                           `json:"should_remove_source_branch,omitempty"`
	ForceRemoveSourceBranch     bool                           `json:"force_remove_source_branch"`
	Reference                   string                         `json:"reference"`
	References                  GitlabReferencesJson           `json:"references"`
	WebURL                      string                         `json:"web_url"`
	TimeStats                   GitlabTimeStatsJson            `json:"time_stats"`
	Squash                      bool                           `json:"squash"`
	TaskCompletionStatus        GitlabTaskCompletionStatusJson `json:"task_completion_status"`
	HasConflicts                bool                           `json:"has_conflicts"`
	BlockingDiscussionsResolved bool                           `json:"blocking_discussions_resolved"`
	ApprovalsBeforeMerge        string                         `json:"approvals_before_merge,omitempty"`
}

type GitlabRepositoryJson struct {
	ID                                        int                                 `json:"id"`
	Description                               string                              `json:"description"`
	Name                                      string                              `json:"name"`
	NameWithNamespace                         string                              `json:"name_with_namespace"`
	Path                                      string                              `json:"path"`
	PathWithNamespace                         string                              `json:"path_with_namespace"`
	CreatedAt                                 time.Time                           `json:"created_at"`
	DefaultBranch                             string                              `json:"default_branch"`
	TagList                                   []string                            `json:"tag_list,omitempty"`
	Topics                                    []string                            `json:"topics,omitempty"`
	SSHURLToRepo                              string                              `json:"ssh_url_to_repo"`
	HTTPURLToRepo                             string                              `json:"http_url_to_repo"`
	WebURL                                    string                              `json:"web_url"`
	ReadmeURL                                 string                              `json:"readme_url,omitempty"`
	AvatarURL                                 string                              `json:"avatar_url,omitempty"`
	ForksCount                                int                                 `json:"forks_count"`
	StarCount                                 int                                 `json:"star_count"`
	LastActivityAt                            time.Time                           `json:"last_activity_at"`
	Namespace                                 GitlabNamespaceJson                 `json:"namespace"`
	ContainerRegistryImagePrefix              string                              `json:"container_registry_image_prefix"`
	Links                                     GitlabLinksJson                     `json:"_links"`
	PackagesEnabled                           bool                                `json:"packages_enabled"`
	EmptyRepo                                 bool                                `json:"empty_repo"`
	Archived                                  bool                                `json:"archived"`
	Visibility                                string                              `json:"visibility"`
	Owner                                     GitlabUserJson                      `json:"owner"`
	ResolveOutdatedDiffDiscussions            bool                                `json:"resolve_outdated_diff_discussions"`
	ContainerExpirationPolicy                 GitlabContainerExpirationPolicyJson `json:"container_expiration_policy"`
	IssuesEnabled                             bool                                `json:"issues_enabled"`
	MergeRequestsEnabled                      bool                                `json:"merge_requests_enabled"`
	WikiEnabled                               bool                                `json:"wiki_enabled"`
	JobsEnabled                               bool                                `json:"jobs_enabled"`
	SnippetsEnabled                           bool                                `json:"snippets_enabled"`
	ContainerRegistryEnabled                  bool                                `json:"container_registry_enabled"`
	ServiceDeskEnabled                        bool                                `json:"service_desk_enabled"`
	ServiceDeskAddress                        string                              `json:"service_desk_address,omitempty"`
	CanCreateMergeRequestIn                   bool                                `json:"can_create_merge_request_in"`
	IssuesAccessLevel                         string                              `json:"issues_access_level"`
	RepositoryAccessLevel                     string                              `json:"repository_access_level"`
	MergeRequestsAccessLevel                  string                              `json:"merge_requests_access_level"`
	ForkingAccessLevel                        string                              `json:"forking_access_level"`
	WikiAccessLevel                           string                              `json:"wiki_access_level"`
	BuildsAccessLevel                         string                              `json:"builds_access_level"`
	SnippetsAccessLevel                       string                              `json:"snippets_access_level"`
	PagesAccessLevel                          string                              `json:"pages_access_level"`
	OperationsAccessLevel                     string                              `json:"operations_access_level"`
	AnalyticsAccessLevel                      string                              `json:"analytics_access_level"`
	ContainerRegistryAccessLevel              string                              `json:"container_registry_access_level"`
	SecurityAndComplianceAccessLevel          string                              `json:"security_and_compliance_access_level"`
	EmailsDisabled                            bool                                `json:"emails_disabled,omitempty"`
	SharedRunnersEnabled                      bool                                `json:"shared_runners_enabled"`
	LfsEnabled                                bool                                `json:"lfs_enabled"`
	CreatorID                                 int                                 `json:"creator_id"`
	ImportURL                                 string                              `json:"import_url,omitempty"`
	Importtype                                string                              `json:"import_type,omitempty"`
	ImportStatus                              string                              `json:"import_status"`
	OpenIssuesCount                           int                                 `json:"open_issues_count"`
	CiDefaultGitDepth                         int                                 `json:"ci_default_git_depth"`
	CiForwardDeploymentEnabled                bool                                `json:"ci_forward_deployment_enabled"`
	CiJobTokenScopeEnabled                    bool                                `json:"ci_job_token_scope_enabled"`
	CiSeparatedCaches                         bool                                `json:"ci_separated_caches"`
	PublicJobs                                bool                                `json:"public_jobs"`
	BuildTimeout                              int                                 `json:"build_timeout"`
	AutoCancelPendingPipelines                string                              `json:"auto_cancel_pending_pipelines"`
	BuildCoverageRegex                        string                              `json:"build_coverage_regex,omitempty"`
	CiConfigPath                              string                              `json:"ci_config_path,omitempty"`
	SharedWithGroups                          []string                            `json:"shared_with_groups,omitempty"`
	OnlyAllowMergeIfPipelineSucceeds          bool                                `json:"only_allow_merge_if_pipeline_succeeds"`
	AllowMergeOnSkippedPipeline               string                              `json:"allow_merge_on_skipped_pipeline,omitempty"`
	RestrictUserDefinedVariables              bool                                `json:"restrict_user_defined_variables"`
	RequestAccessEnabled                      bool                                `json:"request_access_enabled"`
	OnlyAllowMergeIfAllDiscussionsAreResolved bool                                `json:"only_allow_merge_if_all_discussions_are_resolved"`
	RemoveSourceBranchAfterMerge              bool                                `json:"remove_source_branch_after_merge"`
	PrintingMergeRequestLinkEnabled           bool                                `json:"printing_merge_request_link_enabled"`
	MergeMethod                               string                              `json:"merge_method"`
	SquashOption                              string                              `json:"squash_option"`
	SuggestionCommitMessage                   string                              `json:"suggestion_commit_message,omitempty"`
	MergeCommitTemplate                       string                              `json:"merge_commit_template,omitempty"`
	SquashCommitTemplate                      string                              `json:"squash_commit_template,omitempty"`
	AutoDevopsEnabled                         bool                                `json:"auto_devops_enabled"`
	AutoDevopsDeployStrategy                  string                              `json:"auto_devops_deploy_strategy"`
	AutocloseReferencedIssues                 bool                                `json:"autoclose_referenced_issues"`
	RepositoryStorage                         string                              `json:"repository_storage"`
	KeepLatestArtifact                        bool                                `json:"keep_latest_artifact"`
	RunnerTokenExpirationInterval             string                              `json:"runner_token_expiration_interval,omitempty"`
	RequirementsEnabled                       bool                                `json:"requirements_enabled"`
	RequirementsAccessLevel                   string                              `json:"requirements_access_level"`
	SecurityAndComplianceEnabled              bool                                `json:"security_and_compliance_enabled"`
	ComplianceFrameworks                      []string                            `json:"compliance_frameworks,omitempty"`
	Permissions                               GitlabPermissionsJson               `json:"permissions"`
}
type GitlabBranchJson struct {
	Name               string           `json:"name"`
	Commit             GitlabCommitJson `json:"commit"`
	Merged             bool             `json:"merged"`
	Protected          bool             `json:"protected"`
	DevelopersCanPush  bool             `json:"developers_can_push"`
	DevelopersCanMerge bool             `json:"developers_can_merge"`
	CanPush            bool             `json:"can_push"`
	Default            bool             `json:"default"`
	WebURL             string           `json:"web_url"`
}
type GitlabCommitJson struct {
	ID             string    `json:"id"`
	ShortID        string    `json:"short_id"`
	CreatedAt      time.Time `json:"created_at"`
	ParentIds      []string  `json:"parent_ids"`
	Title          string    `json:"title"`
	Message        string    `json:"message"`
	AuthorName     string    `json:"author_name"`
	AuthorEmail    string    `json:"author_email"`
	AuthoredDate   time.Time `json:"authored_date"`
	CommitterName  string    `json:"committer_name"`
	CommitterEmail string    `json:"committer_email"`
	CommittedDate  time.Time `json:"committed_date"`
	Trailers       struct {
	} `json:"trailers"`
	WebURL string `json:"web_url"`
}
type GitlabDiffJson struct {
	OldPath     string `json:"old_path"`
	NewPath     string `json:"new_path"`
	AMode       string `json:"a_mode"`
	BMode       string `json:"b_mode"`
	NewFile     bool   `json:"new_file"`
	RenamedFile bool   `json:"renamed_file"`
	DeletedFile bool   `json:"deleted_file"`
	Diff        string `json:"diff"`
}
type GitlabReferencesJson struct {
	Short    string `json:"short"`
	Relative string `json:"relative"`
	Full     string `json:"full"`
}
type GitlabTimeStatsJson struct {
	TimeEstimate        int    `json:"time_estimate"`
	TotalTimeSpent      int    `json:"total_time_spent"`
	HumanTimeEstimate   string `json:"human_time_estimate,omitempty"`
	HumanTotalTimeSpent string `json:"human_total_time_spent,omitempty"`
}
type GitlabTaskCompletionStatusJson struct {
	Count          int `json:"count"`
	CompletedCount int `json:"completed_count"`
}
type GitlabNamespaceJson struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Kind      string `json:"kind"`
	FullPath  string `json:"full_path"`
	ParentID  string `json:"parent_id,omitempty"`
	AvatarURL string `json:"avatar_url"`
	WebURL    string `json:"web_url"`
}
type GitlabLinksJson struct {
	Self          string `json:"self"`
	Issues        string `json:"issues"`
	MergeRequests string `json:"merge_requests"`
	RepoBranches  string `json:"repo_branches"`
	Labels        string `json:"labels"`
	Events        string `json:"events"`
	Members       string `json:"members"`
}
type GitlabContainerExpirationPolicyJson struct {
	Cadence       string    `json:"cadence"`
	Enabled       bool      `json:"enabled"`
	KeepN         int       `json:"keep_n"`
	OlderThan     string    `json:"older_than"`
	NameRegex     string    `json:"name_regex"`
	NameRegexKeep string    `json:"name_regex_keep,omitempty"`
	NextRunAt     time.Time `json:"next_run_at"`
}
type GitlabPermissionsJson struct {
	ProjectAccess string `json:"project_access,omitempty"`
	GroupAccess   string `json:"group_access,omitempty"`
}
