package blog

import (
	"embed"
	"io/fs"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

//go:embed posts/*
var PostsFS embed.FS

// DSCache is a DateSortable slice of posts
// which only contains the metadata of a post.
//
// This allows us to quickly read out date ordered posts
// without walking the embeded filesystem more then once.
var DSCache DateSortable

func init() {
	var err error
	DSCache, err = NewDSCache()
	if err != nil {
		panic("could not create DSCache: " + err.Error())
	}
}

func NewDSCache() (DateSortable, error) {
	sorted := DateSortable{}
	err := fs.WalkDir(PostsFS, "posts", func(p string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(d.Name()) != ".post" {
			return nil
		}
		f, err := PostsFS.Open(p)
		if err != nil {
			return err
		}

		var post Post
		err = yaml.NewDecoder(f).Decode(&post)
		if err != nil {
			return err
		}
		// .empty is a trick to embed an "empty" posts directory
		// as embed fs api require at least a file inside a dir
		// its embedding. we will just ignore it.
		if post.Title == "_empty" {
			return nil
		}

		sorted = append(sorted, Post{
			Path:    p,
			Title:   post.Title,
			Summary: post.Summary,
			Date:    post.Date,
			Hero:    post.Hero,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(sorted) > 0 {
		sort.Sort(sorted)
	}
	return sorted, nil
}
