package blog

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func WebHandler(appPaths []string) http.HandlerFunc {
	const (
		webPath  = "web"
		blogPath = "blog"
	)
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}

		// any requested web files will expect to be
		// hosted at root, WebFS however embeds files
		// rooted at "web/" so add the web root.
		p := path.Join(webPath, r.URL.Path)

		if p == webPath {
			p = path.Join(webPath, "index.html")
		}

		// if the incoming request is a path defined
		// by the front-end application, serve index.html
		for _, appPath := range appPaths {
			if filepath.HasPrefix(r.URL.Path, appPath) {
				p = path.Join(webPath, "index.html")
			}
		}

		f, err := WebFS.Open(p)
		var fsErr *fs.PathError
		switch {
		case errors.As(err, &fsErr):
			http.Error(w, "not found", http.StatusNotFound)
			return
		case err != nil:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		io.Copy(w, f)
	}
}

func SummaryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}

		var lim int
		var err error
		tmp := r.URL.Query().Get("limit")
		if tmp != "" {
			lim, err = strconv.Atoi(tmp)
			if err != nil {
				http.Error(w, "could not parse limit param: "+err.Error(), http.StatusBadRequest)
				return
			}
		}

		var summaries []Post
		switch {
		case lim == 0:
			summaries = DSCache[:]
		case lim > len(DSCache):
			lim = len(DSCache)
			summaries = DSCache[:lim]
		default:
			summaries = DSCache[:lim]
		}

		err = json.NewEncoder(w).Encode(summaries)
		if err != nil {
			http.Error(w, "failed serializing: "+err.Error(), http.StatusInternalServerError)
		}
	}

}

func PostsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}

		post := r.URL.Path
		post = strings.Trim(post, "/")
		if post == "" || post == "posts" {
			http.Error(w, "no asset provided in path", http.StatusBadRequest)
		}

		f, err := PostsFS.Open(post)
		var fsErr *fs.PathError
		switch {
		case errors.As(err, &fsErr):
			http.Error(w, "not found", http.StatusNotFound)
			return
		case err != nil:
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		// if its not a .post file, its an asset.
		// so just serve it.
		if filepath.Ext(post) != ".post" {
			io.Copy(w, f)
			return
		}

		var markdown Post
		err = yaml.NewDecoder(f).Decode(&markdown)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		w.Header().Add("Content-Type", "text/markdown; charset=UTF-8")
		w.Write([]byte(markdown.MarkDown.Value))
		return
	}
}
