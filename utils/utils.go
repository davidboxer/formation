package utils

import v1 "k8s.io/api/core/v1"

// MergeResourceRequirements Merge src onto dest, overwriting any existing values in dest
func MergeResourceRequirements(dest, src v1.ResourceRequirements) *v1.ResourceRequirements {
	rst := dest.DeepCopy()
	if rst.Requests == nil {
		rst.Requests = make(v1.ResourceList)
	}
	if rst.Limits == nil {
		rst.Limits = make(v1.ResourceList)
	}
	for k, v := range src.Requests {
		rst.Requests[k] = v
	}
	for k, v := range src.Limits {
		rst.Limits[k] = v
	}
	return rst
}

// MergeAffinity Merge new onto old, overwriting any existing values in old
func MergeAffinity(dest, src v1.Affinity) *v1.Affinity {
	rst := dest.DeepCopy()
	skipPodAffinity := src.PodAffinity == nil
	skipPodAntiAffinity := src.PodAntiAffinity == nil
	skipNodeAffinity := src.NodeAffinity == nil

	if rst.PodAffinity == nil && !skipPodAffinity {
		rst.PodAffinity = src.PodAffinity
		skipPodAffinity = true
	}
	if rst.PodAntiAffinity == nil && !skipPodAntiAffinity {
		rst.PodAntiAffinity = src.PodAntiAffinity
		skipPodAntiAffinity = true
	}
	if rst.NodeAffinity == nil && !skipNodeAffinity {
		rst.NodeAffinity = src.NodeAffinity
		skipNodeAffinity = true
	}
	//If PodAffinity is not skip, both are non nil.
	if !skipPodAffinity {
		rst.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
			rst.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
			src.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution...)

		rst.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(
			rst.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
			src.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution...)
	}
	if !skipPodAntiAffinity {
		rst.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
			rst.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
			src.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution...)

		rst.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(
			rst.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
			src.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution...)
	}
	if !skipNodeAffinity {
		rst.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
			rst.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
			src.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution...)

		skipNodeAffinity2 := src.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil
		if rst.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil && !skipNodeAffinity2 {
			rst.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = src.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
		}

		if !skipNodeAffinity2 {
			rst.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = append(
				rst.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
				src.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms...)
		}

	}
	return rst
}

// MergeTolerations Merge new onto old, overwriting any existing values in old
func MergeTolerations(dest, src []v1.Toleration) []v1.Toleration {
	// Make sure we don't modify the original
	rst := make([]v1.Toleration, len(dest))
	copy(rst, dest)
	// Merge the src tolerations into the dest tolerations
	for _, t := range src {
		found := false
		for i, d := range rst {
			// If the toleration already exists, overwrite it
			if t.Key == d.Key {
				rst[i] = t
				found = true
				break
			}
		}
		if !found {
			rst = append(rst, t)
		}
	}
	return rst
}

// MergeTopologySpreadConstraints
func MergeTopologySpreadConstraints(dest, src []v1.TopologySpreadConstraint) []v1.TopologySpreadConstraint {
	// Make sure we don't modify the original
	rst := make([]v1.TopologySpreadConstraint, len(dest))
	copy(rst, dest)
	// Merge the src tolerations into the dest tolerations
	for _, t := range src {
		found := false
		for i, d := range rst {
			// If the toleration already exists, overwrite it
			if t.TopologyKey == d.TopologyKey {
				rst[i] = t
				found = true
				break
			}
		}
		if !found {
			rst = append(rst, t)
		}
	}
	return rst
}
