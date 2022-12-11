package apps

import (
	"github.com/Doout/formation/builder"
	"github.com/Doout/formation/internal/utils"
	"github.com/Doout/formation/resources/apps"
	"github.com/Doout/formation/types"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeploymentBuilder struct {
	*PodBuilder
	Deployment *appsv1.Deployment
}

func NewDeploymentBuilder(name string) DeploymentBuilder {
	obj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Name:        name,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: make(map[string]string),
			},
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{},
			},
		},
	}
	db := DeploymentBuilder{
		PodBuilder: &PodBuilder{
			Builder: builder.Builder{
				Object: obj,
			},
			Spec: &obj.Spec.Template.Spec,
		},
		Deployment: obj,
	}
	return db
}

func (d *DeploymentBuilder) DeepCopy() *DeploymentBuilder {
	deployCopy := d.Deployment.DeepCopy()
	return &DeploymentBuilder{
		PodBuilder: &PodBuilder{
			Builder: builder.Builder{
				Object: deployCopy,
			},
			Spec: &deployCopy.Spec.Template.Spec,
		},
		Deployment: deployCopy,
	}
}

func (d *DeploymentBuilder) AddMatchLabel(key, value string) {
	//Check if MatchLabels nill, if so create it
	if d.Deployment.Spec.Selector.MatchLabels == nil {
		d.Deployment.Spec.Selector.MatchLabels = make(map[string]string)
	}
	//Check if Template Labels nill, if so create it
	if d.Deployment.Spec.Template.Labels == nil {
		d.Deployment.Spec.Template.Labels = make(map[string]string)
	}

	d.Deployment.Spec.Selector.MatchLabels[key] = value
	d.Deployment.Spec.Template.Labels[key] = value
}

func (d *DeploymentBuilder) AddMatchLabels(labels map[string]string) {
	//Check if MatchLabels nill, if so create it
	if d.Deployment.Spec.Selector.MatchLabels == nil {
		d.Deployment.Spec.Selector.MatchLabels = make(map[string]string)
	}
	//Check if Template Labels nill, if so create it
	if d.Deployment.Spec.Template.Labels == nil {
		d.Deployment.Spec.Template.Labels = make(map[string]string)
	}
	for k, v := range labels {
		d.Deployment.Spec.Selector.MatchLabels[k] = v
		d.Deployment.Spec.Template.Labels[k] = v
	}
}

func (d *DeploymentBuilder) SetReplicas(replicas int) {
	d.Deployment.Spec.Replicas = utils.Pointer(int32(replicas))
}

// CreateResource Create the interface to the Formation controller
func (builder *DeploymentBuilder) CreateResource() types.Resource {
	builder.Deployment.Labels = builder.Labels()
	builder.Deployment.Annotations = builder.Annotations()
	return apps.NewDeployment(builder.Deployment.Name, builder.Deployment)
}

func (d DeploymentBuilder) ToResource() types.Resource {
	return apps.NewDeployment(d.Deployment.Name, d.Deployment)
}
