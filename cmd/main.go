package main

import (
	"os"

	"github.com/rhuss/alfred-google-tasks-workflow/internal/alfred"
)

func main() {
	wf := alfred.NewWorkflow()
	wf.Run(os.Args[1:])
}
