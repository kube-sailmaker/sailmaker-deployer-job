package job

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kube-sailmaker/sailmaker-deployer-job/opts"
	"github.com/kube-sailmaker/sailmaker-deployer-job/utils"
	"github.com/kube-sailmaker/template-gen/entry"
	gmodel "github.com/kube-sailmaker/template-gen/model"
	"log"
)

func Process(config *opts.JobConfig) {
	depSummary, err := generateManifests(config)

	if err != nil {
		log.Fatalf("error generating template %v", err)
	}
	err = applyOutput(depSummary, config.OutputLocation)
	if err != nil {
		log.Fatalf("error applying manifests [%v]", err)
	}
}

func generateManifests(config *opts.JobConfig) (*gmodel.DeploymentItemSummary, error) {
	buff := bytes.NewBufferString(config.Payload)
	var appSpec gmodel.AppSpec
	err := json.NewDecoder(buff).Decode(&appSpec)
	if err != nil {
		log.Fatalf("error parsing release payload %v", err)
	}
	appSpec.Environment = "test"
	appSpec.ReleaseName = appSpec.Namespace
	summary, err := entry.TemplateGenerator(&appSpec, config.AppsLocation, config.ResourcesLocation, config.OutputLocation)
	return summary, err
}

func applyOutput(depSummary *gmodel.DeploymentItemSummary, output string) error {
	log.Println("applying manifests from ", output)

	folders, err := utils.WalkPath(output)
	if err != nil {
		return err
	}

	var lastError error = nil
	var applied = make([]string, 0)
	total := len(depSummary.Items)
	log.Printf("Applying a total of %d apps\n", total)
	for _, item := range depSummary.Items {
		log.Println("-------------------------------------------")
		log.Printf("Applying %d items for %s in %s\n", folders[item.Name], item.Name, item.Path)
		err := utils.ExecuteAndDisplay("kubectl", []string{"apply", "-f", item.Path})
		if err != nil {
			lastError = fmt.Errorf("error applying %s in %s, cause: %s", item.Name, item.Path, err.Error())
			break
		}
		log.Println("-------------------------------------------")
		applied = append(applied, item.Name)
	}
	log.Printf("Applied %d out of %d items\n", len(applied), total)

	return lastError
}
