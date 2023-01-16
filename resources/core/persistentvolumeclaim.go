package core

import (
	"context"
	"github.com/Doout/formation/types"
	v1 "k8s.io/api/core/v1"
	types2 "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PersistentVolumeClaim struct {
	pvc *v1.PersistentVolumeClaim
}

func NewPersistentVolumeClaim(pvc *v1.PersistentVolumeClaim) *PersistentVolumeClaim {
	return &PersistentVolumeClaim{
		pvc: pvc,
	}
}

func (p *PersistentVolumeClaim) Type() string           { return "persistentvolumeclaim" }
func (p *PersistentVolumeClaim) Name() string           { return p.pvc.Name }
func (p *PersistentVolumeClaim) Runtime() client.Object { return &v1.PersistentVolumeClaim{} }
func (p *PersistentVolumeClaim) Create() (client.Object, error) {
	if p.pvc.Annotations == nil {
		p.pvc.Annotations = make(map[string]string)
	}
	p.pvc.Annotations[types.UpdateKey] = "disabled"
	return p.pvc, nil
}

// func (c *Job) Converged(ctx context.Context, cli client.Client, namespace string) (bool, error) {
func (p *PersistentVolumeClaim) Converged(ctx context.Context, cli client.Client, namespace string) (bool, error) {
	// Check if the PersistentVolumeClaim Bound
	pvc := &v1.PersistentVolumeClaim{}
	err := cli.Get(ctx, types2.NamespacedName{Name: p.pvc.Name, Namespace: namespace}, pvc)
	if err != nil {
		return false, err
	}
	if pvc.Status.Phase == v1.ClaimBound {
		return true, nil
	}
	return false, nil
}
