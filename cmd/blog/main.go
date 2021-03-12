package main

import (
	"fmt"
	"os"

	"github.com/ldelossa/blog/cmd/blog/internal/serve"
)

const usage = `The goblog command line serves two purposes.
First it may act as an http server, serving assets and blog posts.
Secondly it helps you write and format blog posts. 

The command is split into subcommands, each containing their own help content.

goblog serve  - serve your blog posts, assests, and web root over http
`

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Error: subcommand required\n\n")
		fmt.Println(usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "serve":
		// serve blocks and performs its own
		// os.Exit calls.
		serve.Serve()
	default:
		fmt.Printf("Error: unrecognized subcommand: %s\n", os.Args[1])
		fmt.Println(usage)
	}
}
