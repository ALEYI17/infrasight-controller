package v1alpha1

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

type Relevant struct {
	Image         string
	NodeSelector  map[string]string
	Tolerations   []corev1.Toleration
	Resources     corev1.ResourceRequirements
	RunPrivileged bool
	EnabledProbes []string
	ServerAddress string
	ServerPort    string
  PrometheusPort string
}

func ExtractRelevantSpec(podSpec corev1.PodSpec) Relevant {
	envVar := podSpec.Containers[0].Env
	var envTracer, envServAddr, envAddrPort, envPromPort string
	for _, env := range envVar {
		if env.Name == "TRACER" {
			envTracer = env.Value
		}
		if env.Name == "SERVER_ADDR" {
			envServAddr = env.Value
		}
		if env.Name == "SERVER_PORT" {
			envAddrPort = env.Value
		}
    if env.Name == "PROMETHEUS_PORT"{
      envPromPort = env.Value
    }
	}
	traces := strings.Split(envTracer, ",")

	for i := range traces {
		traces[i] = strings.TrimSpace(traces[i])
	}

	return Relevant{
		Image:         podSpec.Containers[0].Image,
		NodeSelector:  podSpec.NodeSelector,
		Tolerations:   podSpec.Tolerations,
		Resources:     podSpec.Containers[0].Resources,
		RunPrivileged: podSpec.Containers[0].SecurityContext != nil && podSpec.Containers[0].SecurityContext.Privileged != nil && *podSpec.Containers[0].SecurityContext.Privileged,
		EnabledProbes: traces,
		ServerAddress: envServAddr,
		ServerPort:    envAddrPort,
    PrometheusPort: envPromPort,
	}
}
