package apps

import (
	"context"
	"github.com/Doout/formation/resources/common"
	v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StatefulSet struct {
	*common.SimpleResource[*v1.StatefulSet]
	WaitForConverged bool
}

func NewStatefulSet(statefulSet *v1.StatefulSet) *StatefulSet {
	return &StatefulSet{
		SimpleResource:   common.NewSimpleResource("statefulset", statefulSet),
		WaitForConverged: true,
	}
}

func (c *StatefulSet) Converged(ctx context.Context, cli client.Client, namespace string) (bool, error) {
	if !c.WaitForConverged {
		return true, nil
	}

	statefulSet := &v1.StatefulSet{}
	err := cli.Get(ctx, client.ObjectKey{Name: c.Obj.Name, Namespace: namespace}, statefulSet)
	if err != nil {
		return false, err
	}
	if statefulSet.Status.ReadyReplicas == statefulSet.Status.Replicas {
		return true, nil
	}
	return false, nil
}
