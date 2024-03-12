package apps

import (
	"context"

	"github.com/davidboxer/formation/resources/common"
	v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Deployment struct {
	*common.SimpleResource[*v1.Deployment]
	WaitForConverged bool
}

func NewDeployment(deployment *v1.Deployment) *Deployment {
	return &Deployment{
		SimpleResource:   common.NewSimpleResource("deployment", deployment),
		WaitForConverged: true,
	}
}

func (c *Deployment) Converged(ctx context.Context, cli client.Client, namespace string) (bool, error) {
	if !c.WaitForConverged {
		return true, nil
	}
	deployment := &v1.Deployment{}
	err := cli.Get(ctx, client.ObjectKey{Name: c.Obj.Name, Namespace: namespace}, deployment)
	if err != nil {
		return false, err
	}
	if deployment.Status.ReadyReplicas == deployment.Status.Replicas {
		return true, nil
	}
	return false, nil
}
