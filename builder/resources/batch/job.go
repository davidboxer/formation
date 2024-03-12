package batch

import (
	"github.com/davidboxer/formation/builder"
	"github.com/davidboxer/formation/builder/resources/apps"
	"github.com/davidboxer/formation/resources/batch"
	"github.com/davidboxer/formation/types"
	v1 "k8s.io/api/batch/v1"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type JobBuilder struct {
	*apps.PodBuilder
	Job *v1.Job
}

func NewJobBuilder(name string) *JobBuilder {
	obj := &v1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Name:        name,
		},
		Spec: v1.JobSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: make(map[string]string),
			},
			Template: v12.PodTemplateSpec{
				Spec: v12.PodSpec{
					RestartPolicy: v12.RestartPolicyOnFailure,
				},
			},
		},
	}
	return &JobBuilder{
		PodBuilder: &apps.PodBuilder{
			ConvergedGroup: &types.ConvergedGroup{},
			Builder: builder.Builder{
				Object: obj,
				Name:   name,
			},
			Spec: &obj.Spec.Template.Spec,
		},
		Job: obj,
	}
}

func (d *JobBuilder) AddMatchLabel(key, value string) {
	//Check if Template Labels nill, if so create it
	if d.Job.Spec.Template.Labels == nil {
		d.Job.Spec.Template.Labels = make(map[string]string)
	}

	d.Job.Spec.Template.Labels[key] = value
}

func (d *JobBuilder) AddMatchLabels(labels map[string]string) {
	//Check if Template Labels nill, if so create it
	if d.Job.Spec.Template.Labels == nil {
		d.Job.Spec.Template.Labels = make(map[string]string)
	}
	for k, v := range labels {
		d.Job.Spec.Template.Labels[k] = v
	}
}

// ToResource Create the interface to the Formation controller
func (builder *JobBuilder) ToResource() types.Resource {
	builder.Job.Labels = builder.Labels()
	builder.Job.Annotations = builder.Annotations()
	builder.Job.Name = builder.Name
	a := batch.NewJob(builder.Job)
	a.SetConvergedGroupID(builder.GetConvergedGroupID())
	return a
}
