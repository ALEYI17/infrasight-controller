package v1alpha1

import corev1 "k8s.io/api/core/v1"

type Relevant struct {
	Image         string
	NodeSelector  map[string]string
	Tolerations   []corev1.Toleration
	Resources     corev1.ResourceRequirements
	RunPrivileged bool
}

func ExtractRelevantSpec(podSpec corev1.PodSpec) Relevant {
	return Relevant{
		Image:         podSpec.Containers[0].Image,
		NodeSelector:  podSpec.NodeSelector,
		Tolerations:   podSpec.Tolerations,
		Resources:     podSpec.Containers[0].Resources,
		RunPrivileged: podSpec.Containers[0].SecurityContext != nil && podSpec.Containers[0].SecurityContext.Privileged != nil && *podSpec.Containers[0].SecurityContext.Privileged,
	}
}
