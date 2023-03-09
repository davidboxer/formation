package batch

import (
	"context"
	"github.com/Doout/formation/types"
	vv1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CronJob struct {
	*types.ConvergedGroup
	name             string
	cronJob          *vv1.CronJob
	WaitForConverged bool
	DisableUpdate    bool
}

func NewCronJob(name string, cronJob *vv1.CronJob) *CronJob {
	return &CronJob{name: name, cronJob: cronJob, WaitForConverged: true}
}

func (c *CronJob) Type() string           { return "cronjob" }
func (c *CronJob) Name() string           { return c.name }
func (c *CronJob) Runtime() client.Object { return &vv1.CronJob{} }
func (c *CronJob) Create() (client.Object, error) {
	if c.DisableUpdate {
		if c.cronJob.Annotations == nil {
			c.cronJob.Annotations = make(map[string]string)
		}
		c.cronJob.Annotations[types.UpdateKey] = "disabled"
	}
	return c.cronJob, nil
}
func (c *CronJob) Converged(ctx context.Context, cli client.Client, namespace string) (bool, error) {
	if !c.WaitForConverged {
		return true, nil
	}

	cronJob := &vv1.CronJob{}
	err := cli.Get(ctx, client.ObjectKey{Name: c.name, Namespace: namespace}, cronJob)
	if err != nil {
		return false, err
	}
	//Check if any active jobs are running, if any is found, return false
	if len(cronJob.Status.Active) > 0 {
		return false, nil
	}
	return true, nil
}
