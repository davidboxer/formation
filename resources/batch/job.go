package batch

import (
	"context"

	"github.com/davidboxer/formation/resources/common"
	"github.com/davidboxer/formation/types"
	v1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Job struct {
	*common.SimpleResource[*v1.Job]
	WaitForConverged bool
}

func NewJob(job *v1.Job) *Job {
	return &Job{
		SimpleResource:   common.NewSimpleResource("job", job),
		WaitForConverged: true,
	}
}
func (c *Job) Create() (client.Object, error) {
	//Once a Job is created, it is not possible to update it.
	if c.Obj.Annotations == nil {
		c.Obj.Annotations = make(map[string]string)
	}
	c.Obj.Annotations[types.UpdateKey] = "disabled"
	return c.Obj, nil
}

func (c *Job) Update(ctx context.Context, fromApiServer runtime.Object) error {
	return nil
}

func (c *Job) Converged(ctx context.Context, cli client.Client, namespace string) (bool, error) {
	if !c.WaitForConverged {
		return true, nil
	}

	job := &v1.Job{}
	err := cli.Get(ctx, client.ObjectKey{Name: c.Obj.Name, Namespace: namespace}, job)
	if err != nil {
		return false, err
	}
	if job.Status.Succeeded > 0 {
		return true, nil
	}
	return false, nil
}
