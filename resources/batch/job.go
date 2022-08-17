package batch

import (
	"context"
	"github.com/Doout/formation/types"
	v1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Job struct {
	name             string
	job              *v1.Job
	WaitForConverged bool
}

func NewJob(name string, job *v1.Job) *Job {
	return &Job{name: name, job: job, WaitForConverged: true}
}

func (c *Job) Type() string           { return "job" }
func (c *Job) Name() string           { return c.name }
func (c *Job) Runtime() client.Object { return &v1.Job{} }

func (c *Job) Create() (client.Object, error) {
	//Once a Job is created, it is not possible to update it.
	if c.job.Annotations == nil {
		c.job.Annotations = make(map[string]string)
	}
	c.job.Annotations[types.UpdateKey] = "disabled"
	return c.job, nil
}

func (c *Job) Update(ctx context.Context, fromApiServer runtime.Object) error {
	return nil
}

func (c *Job) Converged(ctx context.Context, cli client.Client, namespace string) (bool, error) {
	if !c.WaitForConverged {
		return true, nil
	}

	job := &v1.Job{}
	err := cli.Get(ctx, client.ObjectKey{Name: c.name, Namespace: namespace}, job)
	if err != nil {
		return false, err
	}
	if job.Status.Succeeded > 0 {
		return true, nil
	}
	return false, nil
}
