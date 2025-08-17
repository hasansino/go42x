package main

import (
	"github.com/hasansino/go42x/cmd"
	_ "github.com/hasansino/go42x/cmd/generate"
	_ "github.com/hasansino/go42x/cmd/generate/ai"
	_ "github.com/hasansino/go42x/cmd/version"
)

func main() {
	cmd.Execute()
}
