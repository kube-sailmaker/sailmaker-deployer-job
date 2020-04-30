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
	buff := bytes.NewBufferString(config.Payload)
	var appSpec gmodel.AppSpec
	err := json.NewDecoder(buff).Decode(&appSpec)
	if err != nil {
		log.Fatalf("error parsing release payload %v", err)
	}
	entry.TemplateGenerator(&appSpec, config.AppsLocation, config.ResourcesLocation, config.OutputLocation)
	err = applyOutput(config.OutputLocation)
	if err != nil {
		log.Fatalf("error applying manifests [%v]", err)
	}
}

func applyOutput(output string) error {
	log.Println("applying manifests from ", output)

	folders, err := utils.WalkPath(output)
	if err != nil {
		return err
	}

	var lastError error = nil
	var applied = make([]string, 0)
	total := len(folders)
	log.Printf("Applying a total of %d apps\n", total)
	for folder, items := range folders {
		log.Println("-------------------------------------------")
		log.Printf("Applying %d items in %s\n", items, folder)
		err := utils.ExecuteAndDisplay("kubectl", []string{"apply", "-f", folder})
		if err != nil {
			lastError = fmt.Errorf("error applying %s, cause: %s", folder, err.Error())
			break
		}
		log.Println("-------------------------------------------")
		applied = append(applied, folder)
	}
	log.Printf("Applied %d out of %d items\n", len(applied), total)

	return lastError
}
