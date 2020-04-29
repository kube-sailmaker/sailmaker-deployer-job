package main

import (
	"github.com/kube-sailmaker/sailmaker-deployer-job/job"
	"github.com/kube-sailmaker/sailmaker-deployer-job/opts"
	"os"
)

func main() {
	flags := opts.FlagSets()
	switch os.Args[1] {
	case "deploy":
		job.Process(opts.ProcessDeployment(flags["deploy"]))
	default:
		opts.ExitWithUsage("Unknown command", flags["deploy"])
	}

}
