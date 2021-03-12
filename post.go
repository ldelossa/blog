package blog

import (
	"time"

	"gopkg.in/yaml.v3"
)

// DateSortable is a slice of Posts sortable
// by time.
type DateSortable []Post

func (t DateSortable) Len() int {
	return len(t)
}

func (t DateSortable) Less(i, j int) bool {
	return t[i].Date.After(t[j].Date)
}

func (t DateSortable) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

// Post is a markdown blog post.
//
// The structure provides the markdown contents
// along with some metadata used to summarize the post.
type Post struct {
	// internally used; the path in the embed.FS where the
	// contents the post can be read.
	Path    string    `json:"path" yaml:"-"`
	Hero    string    `json:"hero" yaml:"hero"`
	Title   string    `json:"title" yaml:"title"`
	Summary string    `json:"summary" yaml:"summary"`
	Date    time.Time `json:"date" yaml:"date"`
	// the markdown body of the blog post.
	MarkDown yaml.Node `json:"-" yaml:"mark_down,omitempty"`
}
