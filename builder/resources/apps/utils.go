package apps

import (
	"github.com/davidboxer/formation/types"
	v1 "k8s.io/api/core/v1"
)

// setProbeConfiguration set the probe's parameters based on the configuration.
// If the probe is nil then the configuration is not applied.
func setProbeConfiguration(probe *v1.Probe, config types.ProbeConfiguration) {
	// If the probe is nil, don't configure
	if probe == nil {
		return
	}
	// If the  probe is not enabled, set the probe to nil
	if !config.Enable {
		probe = nil
		return
	}
	// Set the probe's parameters
	if config.InitialDelaySeconds != nil {
		probe.InitialDelaySeconds = *config.InitialDelaySeconds
	}
	if config.TimeoutSeconds != nil {
		probe.TimeoutSeconds = *config.TimeoutSeconds
	}
	if config.PeriodSeconds != nil {
		probe.PeriodSeconds = *config.PeriodSeconds
	}
	if config.SuccessThreshold != nil {
		probe.SuccessThreshold = *config.SuccessThreshold
	}
	if config.FailureThreshold != nil {
		probe.FailureThreshold = *config.FailureThreshold
	}
}
