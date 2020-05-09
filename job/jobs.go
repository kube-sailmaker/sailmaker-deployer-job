package job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kube-sailmaker/sailmaker-deployer-job/k8s/client"
	"github.com/kube-sailmaker/sailmaker-deployer-job/model"
	"github.com/kube-sailmaker/sailmaker-deployer-job/opts"
	"github.com/kube-sailmaker/sailmaker-deployer-job/utils"
	"github.com/kube-sailmaker/template-gen/entry"
	gmodel "github.com/kube-sailmaker/template-gen/model"
	"k8s.io/api/extensions/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
)

func Process(config *opts.JobConfig) {
	depSummary, err := generateManifests(config)

	if err != nil {
		log.Fatalf("error generating template %v", err)
	}
	if depSummary.Namespace == "" {
		log.Fatalf("namespace is required for the release")
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
	envName := os.Getenv("SAILMAKER_ENV")
	if envName == "" {
		envName = "local"
	}
	appSpec.Environment = envName
	appSpec.ReleaseName = appSpec.Namespace
	summary, err := entry.TemplateGenerator(&appSpec, config.AppsLocation, config.ResourcesLocation, config.OutputLocation)
	if summary != nil && summary.Namespace == "" {
		summary.Namespace = appSpec.Namespace
	}
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

	deployables := make(map[string]model.DeploymentStatus, 0)
	for _, it := range depSummary.Items {
		deployName := fmt.Sprintf("%s-%s", depSummary.Namespace, it.Name)
		deployables[deployName] = model.DeploymentStatus{
			Name:      deployName,
			Completed: false,
			Status:    "Incomplete",
		}
	}
	watch := make(chan int)
	go deploymentWatcher(depSummary.Namespace, deployables, watch)

	for _, item := range depSummary.Items {
		log.Println("-------------------------------------------")
		log.Printf("Applying %d items for %s in %s\n", folders[item.Name], item.Name, item.Path)

		err := utils.ExecuteAndDisplay("kubectl", []string{"apply", "-f", item.Path, "-n", depSummary.Namespace})
		if err != nil {
			lastError = fmt.Errorf("error applying %s in %s, cause: %s", item.Name, item.Path, err.Error())
			break
		}
		log.Println("-------------------------------------------")
		applied = append(applied, item.Name)
	}
	log.Printf("Applied %d out of %d items\n", len(applied), total)

	watchResult, ok := <-watch
	if !ok {
		return fmt.Errorf("error updating deployments")
	}
	log.Printf("watched and updated %d deployments", watchResult)
	return lastError
}

func deploymentWatcher(namespace string, deployables map[string]model.DeploymentStatus, w chan int) {
	k8s := client.GetClient()
	deploymentList, _ := k8s.ExtensionsV1beta1().Deployments(namespace).List(context.TODO(), v1.ListOptions{})
	if deploymentList != nil {
		log.Println("Existing Deployments", len(deploymentList.Items))
	}
	watcher, err := k8s.ExtensionsV1beta1().Deployments(namespace).Watch(context.TODO(), v1.ListOptions{})
	if err != nil {
		log.Fatal("error when watching for deployment status", err)
	}
	event := watcher.ResultChan()
	expected := len(deployables)
	actual := 0
	log.Printf("watching in for deployment events in namespace %s\n", namespace)
	for {
		item, ok := <-event
		if !ok {
			break
		}
		deployment := item.Object
		singleDeploymentComplete := false
		switch t := deployment.(type) {
		case *v1beta1.Deployment:
			deploymentRecord, ok := deployables[t.Name]
			if ok {
				ready := t.Status.ReadyReplicas
				available := t.Status.AvailableReplicas
				total := t.Status.Replicas
				requiredReplicas := t.Spec.Replicas
				unavailableReplicas := t.Status.UnavailableReplicas
				if *requiredReplicas == total && total == ready && total == available && unavailableReplicas == 0 {
					actual = actual + 1
					log.Printf("deployment completed for %s. %d out of %d finished.\n", deploymentRecord.Name, actual, expected)
					deploymentRecord.Completed = true
					deploymentRecord.Reason = "condition met."
					deploymentRecord.Status = "Complete"
					deployables[t.Name] = deploymentRecord
					singleDeploymentComplete = true
				}
			}
		default:
			log.Println("Unknown", deployment)
		}
		if singleDeploymentComplete {
			for _, d := range deployables {
				log.Printf("deployment=%s, status=%s\n", d.Name, d.Status)
			}
		}
		if actual == expected {
			break
		}
	}
	watcher.Stop()
	w <- actual
}
