package apps

import (
	"context"
	"github.com/Doout/formation/types"
	v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Deployment struct {
	name             string
	deployment       *v1.Deployment
	WaitForConverged bool
	DisableUpdate    bool
}

// Interface for Deployment for the Formation Controller
func NewDeployment(name string, deployment *v1.Deployment) *Deployment {
	return &Deployment{name: name, deployment: deployment, WaitForConverged: true}
}

func (c *Deployment) Type() string { return "deployment" }

func (c *Deployment) Name() string { return c.name }

func (c *Deployment) Runtime() client.Object { return &v1.Deployment{} }

func (c *Deployment) Create() (client.Object, error) {
	if c.DisableUpdate {
		if c.deployment.Annotations == nil {
			c.deployment.Annotations = make(map[string]string)
		}
		c.deployment.Annotations[types.UpdateKey] = "disabled"
	}
	return c.deployment, nil
}

func (c *Deployment) Converged(ctx context.Context, cli client.Client, namespace string) (bool, error) {
	if !c.WaitForConverged {
		return true, nil
	}
	deployment := &v1.Deployment{}
	err := cli.Get(ctx, client.ObjectKey{Name: c.name, Namespace: namespace}, deployment)
	if err != nil {
		return false, err
	}
	if deployment.Status.ReadyReplicas == deployment.Status.Replicas {
		return true, nil
	}
	return false, nil
}
