package opts

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

type JobConfig struct {
	ResourcesLocation string `json:"resources"`
	AppsLocation      string `json:"apps"`
	Payload           string `json:"releases"`
	OutputLocation    string `json:"output"`
}

func ProcessDeployment(dOpts *flag.FlagSet) *JobConfig {
	err := dOpts.Parse(os.Args[2:])
	if err != nil {
		ExitWithUsage("Parse Error", dOpts)
	}
	resourcesLocation := dOpts.Lookup("resources").Value.String()
	appDef := dOpts.Lookup("apps").Value.String()
	releases := dOpts.Lookup("releases").Value.String()
	output := dOpts.Lookup("output").Value.String()

	if !dOpts.Parsed() {
		ExitWithUsage("Parse Error", dOpts)
	}
	if resourcesLocation == "" || appDef == "" || output == "" {
		ExitWithUsage("Resources, Releases and Output Location required", dOpts)
	}
	var payload = ""
	if releases == "" {
		payloadFromEnv := os.Getenv("RELEASE_PAYLOAD")
		if payloadFromEnv == "" {
			ExitWithUsage("RELEASE_PAYLOAD or --releases is required", dOpts)
		}
		payload = payloadFromEnv
	} else {
		rel, err := ioutil.ReadFile(releases)
		if err != nil {
			ExitWithUsage(fmt.Sprintf("Release File %s could not be read.", releases), dOpts)
		}
		payload = string(rel)
	}
	return &JobConfig{
		ResourcesLocation: resourcesLocation,
		AppsLocation:      appDef,
		Payload:           payload,
		OutputLocation: output,
	}

}

func FlagSets() map[string]*flag.FlagSet {
	dOpts := flag.NewFlagSet("deploy", flag.ExitOnError)
	dOpts.String("resources", "", "Resource Location where infrastructure, mixins and resources are")
	dOpts.String("apps", "", "Location where Application Definitions are")
	dOpts.String("releases", "", "Location where release file is")
	dOpts.String("output", "", "Output Directory for manifests")
	return map[string]*flag.FlagSet{
		"deploy": dOpts,
	}
}

func ExitWithUsage(msg string, opts *flag.FlagSet) {
	fmt.Println(msg)
	opts.Usage()
	os.Exit(1)
}
