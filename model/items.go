package model

type DeploymentStatus struct {
	Name      string `json:"name"`
	Completed bool   `json:"completed"`
	Status    string `json:"status"`
	Reason    string `json:"reason"`
}
