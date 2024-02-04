package apps

import (
	"github.com/Doout/formation/utils"
	"k8s.io/api/core/v1"
	"strconv"
	"strings"
)

type ContainerBuilder struct {
	*v1.Container
}

func NewContainer(name string) *ContainerBuilder {
	return &ContainerBuilder{
		Container: &v1.Container{
			Name:                     name,
			TerminationMessagePath:   "/dev/termination-log",
			TerminationMessagePolicy: v1.TerminationMessageReadFile,
			ImagePullPolicy:          v1.PullIfNotPresent,
			SecurityContext: &v1.SecurityContext{
				AllowPrivilegeEscalation: utils.ToPointer(false),
				Capabilities: &v1.Capabilities{
					Drop: []v1.Capability{"ALL"},
				},
				SeccompProfile: &v1.SeccompProfile{
					Type: "RuntimeDefault",
				},
				RunAsNonRoot: utils.ToPointer(true),
			},
		},
	}
}

func (c *ContainerBuilder) SetCommand(command []string) *ContainerBuilder {
	c.Command = command
	return c
}

func (c *ContainerBuilder) SetResourceRequirements(resources v1.ResourceRequirements) *ContainerBuilder {
	c.Resources = resources
	return c
}

func (c *ContainerBuilder) SetArgs(args []string) *ContainerBuilder {
	c.Args = args
	return c
}

func (c *ContainerBuilder) SetImage(image string) *ContainerBuilder {
	c.Image = image
	return c
}

// SetReadinessProbe Set the readiness probe for the container
func (c *ContainerBuilder) SetReadinessProbe(readinessProbe v1.Probe) *ContainerBuilder {
	c.ReadinessProbe = readinessProbe.DeepCopy()
	return c
}

// SetLivenessProbe Set the liveness probe for the container
func (c *ContainerBuilder) SetLivenessProbe(livenessProbe v1.Probe) *ContainerBuilder {
	c.LivenessProbe = livenessProbe.DeepCopy()
	return c
}

// SetStartupProbe Set the startup probe for the container
func (c *ContainerBuilder) SetStartupProbe(startupProbe v1.Probe) *ContainerBuilder {
	c.StartupProbe = startupProbe.DeepCopy()
	return c
}

func (c *ContainerBuilder) SetSecurityContext(securityContext v1.SecurityContext) *ContainerBuilder {
	c.SecurityContext = securityContext.DeepCopy()
	return c
}

func (c *ContainerBuilder) SetImagePullPolicy(imagePullPolicy v1.PullPolicy) *ContainerBuilder {
	c.ImagePullPolicy = imagePullPolicy
	return c
}

func (c *ContainerBuilder) AddPortsRange(startPort v1.ContainerPort, count, step int) *ContainerBuilder {
	lastPort := startPort.DeepCopy()
	if lastPort == nil {
		return c
	}
	name := lastPort.Name
	step32 := int32(step)
	index := 0
	for count > 0 {
		count -= step
		index += step
		c.AddPorts(true, *lastPort)
		lastPort = lastPort.DeepCopy()
		if lastPort.HostPort != 0 {
			lastPort.HostPort += step32
		}
		lastPort.ContainerPort += step32
		lastPort.Name = name + strconv.Itoa(index)
	}
	return c
}

func (c *ContainerBuilder) AddPorts(overwrite bool, ports ...v1.ContainerPort) *ContainerBuilder {
	if c.Env == nil {
		c.Ports = make([]v1.ContainerPort, 0)
		c.Ports = append(c.Ports, ports...)
		return c
	}

	for _, value := range ports {
		index, exist := c.EnvironmentVariableExist(strings.ToLower(value.Name))
		if !exist {
			c.Ports = append(c.Ports, value)
		} else if overwrite {
			value.DeepCopyInto(&c.Ports[index])
		}
	}
	return c
}

func (c *ContainerBuilder) AddEnvironmentVariable(overwrite bool, envVars ...v1.EnvVar) *ContainerBuilder {
	if c.Env == nil {
		c.Env = make([]v1.EnvVar, 0)
		c.Env = append(c.Env, envVars...)
		return c
	}
	for _, value := range envVars {
		index, exist := c.EnvironmentVariableExist(strings.ToLower(value.Name))
		if !exist {
			c.Env = append(c.Env, value)
		} else if overwrite {
			value.DeepCopyInto(&c.Env[index])
		}
	}
	return c
}

func (c *ContainerBuilder) AddEnvironmentFromSource(overwrite bool, envVars ...v1.EnvFromSource) *ContainerBuilder {
	if c.EnvFrom == nil {
		c.EnvFrom = make([]v1.EnvFromSource, 0)
		c.EnvFrom = append(c.EnvFrom, envVars...)
		return c
	}
	for _, value := range envVars {
		index, exist := c.EnvironmentFromSourceExist(value)
		if !exist {
			c.EnvFrom = append(c.EnvFrom, value)
		} else if overwrite {
			c.EnvFrom[index].ConfigMapRef = nil
			c.EnvFrom[index].SecretRef = nil
			value.DeepCopyInto(&c.EnvFrom[index])
		}
	}
	return c
}

func (c *ContainerBuilder) AddEnvironmentVariable2(name, value string, overwrite bool) *ContainerBuilder {
	env := v1.EnvVar{Name: name, Value: value}
	c.AddEnvironmentVariable(overwrite, env)
	return c
}

func (c *ContainerBuilder) SetRole(role string) *ContainerBuilder {
	env := v1.EnvVar{Name: "ROLE", Value: role}
	c.AddEnvironmentVariable(true, env)
	return c
}

func (c *ContainerBuilder) SetTTY(tty bool) *ContainerBuilder {
	c.TTY = tty
	return c
}

func (c *ContainerBuilder) PortExist(portName string) (int, bool) {
	for index, value := range c.Ports {
		if strings.ToLower(value.Name) == portName {
			return index, true
		}
	}
	return -1, false
}

func (c *ContainerBuilder) EnvironmentVariableExist(envVarName string) (int, bool) {
	for index, value := range c.Env {
		if strings.ToLower(value.Name) == envVarName {
			return index, true
		}
	}
	return -1, false
}

func (c *ContainerBuilder) EnvironmentFromSourceExist(envFrom v1.EnvFromSource) (int, bool) {
	for index, value := range c.EnvFrom {
		if value.SecretRef != nil && envFrom.SecretRef != nil && value.SecretRef.Name == envFrom.SecretRef.Name {
			return index, true
		}
		if value.ConfigMapRef != nil && envFrom.ConfigMapRef != nil && value.ConfigMapRef.Name == envFrom.ConfigMapRef.Name {
			return index, true
		}
	}
	return -1, false
}

func (c *ContainerBuilder) ToContainer() *v1.Container {
	return c.Container
}

func ToContainerBuilder(container *v1.Container) *ContainerBuilder {
	return &ContainerBuilder{
		Container: container,
	}
}
