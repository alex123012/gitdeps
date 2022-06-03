package structs

import "time"

type Commit struct {
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

type Diff struct {
	OldPath     string `json:"old_path"`
	NewPath     string `json:"new_path"`
	AMode       string `json:"a_mode"`
	BMode       string `json:"b_mode"`
	NewFile     bool   `json:"new_file"`
	RenamedFile bool   `json:"renamed_file"`
	DeletedFile bool   `json:"deleted_file"`
	Diff        string `json:"diff"`
}

type Compare struct {
	Commit         Commit   `json:"commit"`
	Commits        []Commit `json:"commits"`
	Diffs          []Diff   `json:"diffs"`
	CompareTimeout bool     `json:"compare_timeout"`
	CompareSameRef bool     `json:"compare_same_ref"`
	WebURL         string   `json:"web_url"`
}
