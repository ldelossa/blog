package blog

import (
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// TestPost confirms a Post structure can
// Marshal and Unmarshal our MarkDown blog post.
func TestPost(t *testing.T) {
	const TestPost = `
---
title: A Test Blog Post
summary: The summary of a test blog post
date: 2021-1-1
mark_down: |
  # A header

  a sentence

  ## Another Header

  another sentence
`
	const TestMarkdown = `# A header

a sentence

## Another Header

another sentence
`
	var post Post
	err := yaml.Unmarshal([]byte(TestPost), &post)
	if err != nil {
		t.Fatalf("failed to marshal test post: %v", err)
	}

	if post.Title != "A Test Blog Post" {
		t.Fatalf("got: %v, want: %v", post.Title, "A Test Blog Post")
	}
	if post.Summary != "The summary of a test blog post" {
		t.Fatalf("got: %v, want: %v", post.Title, "The summary of a test blog post")
	}
	if post.Date.Month() != time.January {
		t.Fatalf("got: %v, want: %v", post.Date.Month(), time.January)
	}
	if post.Date.Day() != 1 {
		t.Fatalf("got: %v, want: %v", post.Date.Day(), 1)
	}
	if post.Date.Year() != 2021 {
		t.Fatalf("got: %v, want: %v", post.Date.Year(), 2021)
	}
	if post.MarkDown != TestMarkdown {
		t.Fatalf("got: %v, want: %v", post.MarkDown, TestMarkdown)
	}
}
