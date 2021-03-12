package serve

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ldelossa/blog"
	"github.com/rs/cors"
)

var fs = flag.NewFlagSet("serve", flag.ExitOnError)

var flags = struct {
	listenAddr *string
}{
	listenAddr: fs.String("l", ":8080", "a <host:port> string where goblog will listen for http requests"),
}

// Serve will launch an http server and begin serving blog posts
// and assets
func Serve() {
	// 0: goblog, 1: server
	fs.Parse(os.Args[2:])

	inter := make(chan os.Signal)
	signal.Notify(inter, os.Interrupt)

	var mux http.ServeMux
	// mux.Handle("/assets/", goblog.AssetHandler())
	mux.Handle("/posts/", blog.PostsHandler())
	mux.Handle("/summaries", blog.SummaryHandler())
	mux.Handle("/", blog.WebHandler([]string{
		"/blog",
		"/archive",
		"/home",
		"/settings",
	}))

	server := &http.Server{
		Addr:    *flags.listenAddr,
		Handler: cors.Default().Handler(&mux),
	}

	ctx, cancel := context.WithCancel(context.Background())
	var httpErr error
	go func() {
		log.Printf("Launching goblog @ %v\n", *flags.listenAddr)
		err := server.ListenAndServe()
		if err != nil {
			httpErr = err
			cancel()
		}
	}()

	select {
	case <-inter:
		log.Printf("Received interupt. Gracefully shutting down server.\n")
		tctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		server.Shutdown(tctx)
	case <-ctx.Done():
		if httpErr != http.ErrServerClosed {
			log.Printf("Received http error: %v\n", httpErr)
			os.Exit(1)
		}
	}
	os.Exit(0)
}
