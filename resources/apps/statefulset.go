package apps

import (
	"context"
	"github.com/Doout/formation/types"
	v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StatefulSet struct {
	*types.ConvergedGroup
	name             string
	statefulSet      *v1.StatefulSet
	WaitForConverged bool
	DisableUpdate    bool
}

func NewStatefulSet(name string, statefulSet *v1.StatefulSet) *StatefulSet {
	return &StatefulSet{name: name, statefulSet: statefulSet, WaitForConverged: true}
}

func (c *StatefulSet) Type() string { return "statefulset" }

func (c *StatefulSet) Name() string { return c.name }

func (c *StatefulSet) Runtime() client.Object { return &v1.StatefulSet{} }

func (c *StatefulSet) Create() (client.Object, error) {
	if c.DisableUpdate {
		if c.statefulSet.Annotations == nil {
			c.statefulSet.Annotations = make(map[string]string)
		}
		c.statefulSet.Annotations[types.UpdateKey] = "disabled"
	}
	return c.statefulSet, nil
}

func (c *StatefulSet) Converged(ctx context.Context, cli client.Client, namespace string) (bool, error) {
	if !c.WaitForConverged {
		return true, nil
	}

	statefulSet := &v1.StatefulSet{}
	err := cli.Get(ctx, client.ObjectKey{Name: c.name, Namespace: namespace}, statefulSet)
	if err != nil {
		return false, err
	}
	if statefulSet.Status.ReadyReplicas == statefulSet.Status.Replicas {
		return true, nil
	}
	return false, nil
}
