package batch

import (
	"context"

	"github.com/davidboxer/formation/resources/common"
	vv1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CronJob struct {
	*common.SimpleResource[*vv1.CronJob]
	WaitForConverged bool
}

func NewCronJob(name string, cronJob *vv1.CronJob) *CronJob {
	return &CronJob{
		SimpleResource: common.NewSimpleResource("cronjob", cronJob),
	}
}

func (c *CronJob) Converged(ctx context.Context, cli client.Client, namespace string) (bool, error) {
	if !c.WaitForConverged {
		return true, nil
	}

	cronJob := &vv1.CronJob{}
	err := cli.Get(ctx, client.ObjectKey{Name: c.Obj.Name, Namespace: namespace}, cronJob)
	if err != nil {
		return false, err
	}
	//Check if any active jobs are running, if any is found, return false
	if len(cronJob.Status.Active) > 0 {
		return false, nil
	}
	return true, nil
}
