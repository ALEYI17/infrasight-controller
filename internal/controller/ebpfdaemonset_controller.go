/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	ebpfv1alpha1 "github.com/ALEYI17/kube-ebpf-monitor/api/v1alpha1"
	"github.com/go-logr/logr"
)

type EbpfDaemonSetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ebpf.monitoring.dev,resources=ebpfdaemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ebpf.monitoring.dev,resources=ebpfdaemonsets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ebpf.monitoring.dev,resources=ebpfdaemonsets/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

func (r *EbpfDaemonSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Manage the CR
	ebpfDs := &ebpfv1alpha1.EbpfDaemonSet{}

	err := r.Get(ctx, req.NamespacedName, ebpfDs)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Ebpf DaemonSet not found , ignoring since it may be deleted")
			return ctrl.Result{}, nil
		}

		log.Error(err, "Failed to get Ebpf Daemonset")
		return ctrl.Result{}, nil
	}

	if len(ebpfDs.Status.Conditions) == 0 {

		patch := client.MergeFrom(ebpfDs.DeepCopy())
		meta.SetStatusCondition(&ebpfDs.Status.Conditions,
			metav1.Condition{Type: "Available", Status: metav1.ConditionUnknown, Reason: "Reconciling", Message: "Starting reconciliation"})

		if err := r.Status().Patch(ctx, ebpfDs, patch); err != nil {
			log.Error(err, "Failed to update Status")
			return ctrl.Result{}, err
		}

		if err := r.Get(ctx, req.NamespacedName, ebpfDs); err != nil {
			log.Error(err, "Error to re-fetch")
			return ctrl.Result{}, err
		}
	}

	// Manage the the DaemonSet
	found := &appsv1.DaemonSet{}
	if err := r.Get(ctx, types.NamespacedName{Name: ebpfDs.Name, Namespace: ebpfDs.Namespace}, found); err != nil && apierrors.IsNotFound(err) {
		ds, err := r.DaemonSetForEbpf(ebpfDs)
		if err != nil {

			log.Error(err, "Failed to define new DaemonSet resource for ebpf")
			patch := client.MergeFrom(ebpfDs.DeepCopy())

			meta.SetStatusCondition(&ebpfDs.Status.Conditions, metav1.Condition{Type: "Available", Status: metav1.ConditionFalse, Reason: "Reconciling", Message: fmt.Sprintf("Failed to create DaemonSet for cr (%s): (%s) ", ebpfDs.Name, err)})

			if err := r.Status().Patch(ctx, ebpfDs, patch); err != nil {
				log.Error(err, "Failed to update Status")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}

		if err := controllerutil.SetOwnerReference(ebpfDs, ds, r.Scheme); err != nil {
			log.Error(err, "Unable to SetOwnerReference")
			return ctrl.Result{}, err
		}
		log.Info("Creating new Daemonset", "Namespace", ds.Namespace, "Name", ds.Name)

		if err := r.Create(ctx, ds); err != nil {
			log.Error(err, "Failed to create new Daemonset", "Namespace:", ds.Namespace, "Name", ds.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Minute}, nil

	} else if err != nil {
		log.Error(err, "Failed to get DaemonSet")
		return ctrl.Result{}, err
	}

	desired, err := r.DaemonSetForEbpf(ebpfDs)
	if err != nil {
		log.Error(err, "Unable to generate desired DaemonSet ")
		return ctrl.Result{}, err
	}

	patch := client.MergeFrom(found.DeepCopy())
	if r.CmpDaemonSets(found, desired, log) {
		log.Info("Spec change detected, updating daemonset")

		err := r.Patch(ctx, found, patch)
		if err != nil {
			log.Error(err, "Error updating daemonset")
			return ctrl.Result{}, err
		}

		latest := &ebpfv1alpha1.EbpfDaemonSet{}
		if err := r.Get(ctx, req.NamespacedName, latest); err != nil {
			log.Error(err, "Failed to re-fetch EbpfDaemonSet before updating status")
			return ctrl.Result{}, err
		}

		patch := client.MergeFrom(latest.DeepCopy())
		meta.SetStatusCondition(&latest.Status.Conditions,
			metav1.Condition{Type: "Available", Status: metav1.ConditionUnknown, Reason: "Reconciling",
				Message: "EbpfDaemonSet is updating"})
		if err := r.Status().Patch(ctx, latest, patch); err != nil {
			log.Error(err, "Failed to update status")
			return ctrl.Result{}, err

		}

		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	latest := &ebpfv1alpha1.EbpfDaemonSet{}
	if err := r.Get(ctx, req.NamespacedName, latest); err != nil {
		log.Error(err, "Failed to re-fetch EbpfDaemonSet before updating status")
		return ctrl.Result{}, err
	}

	patch = client.MergeFrom(latest.DeepCopy())
	meta.SetStatusCondition(&latest.Status.Conditions,
		metav1.Condition{Type: "Available", Status: metav1.ConditionTrue, Reason: "Reconciling",
			Message: "EbpfDaemonSet created successfully"})

	if err := r.Status().Patch(ctx, latest, patch); err != nil {
		log.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EbpfDaemonSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ebpfv1alpha1.EbpfDaemonSet{}).
		Owns(&appsv1.DaemonSet{}).
		Named("ebpfdaemonset").
		Complete(r)
}

func newHostPathType(t string) *corev1.HostPathType {
	typ := corev1.HostPathType(t)
	return &typ
}

func (r *EbpfDaemonSetReconciler) DaemonSetForEbpf(ebpfds *ebpfv1alpha1.EbpfDaemonSet) (*appsv1.DaemonSet, error) {
	label := map[string]string{
		"app": ebpfds.Name,
	}

	volumes := []corev1.Volume{
		{
			Name: "debugfs",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/sys/kernel/debug",
					Type: newHostPathType("Directory"),
				},
			},
		},
		{
			Name: "var-run",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/var/run",
					Type: newHostPathType("Directory"),
				},
			},
		},
		{
			Name: "host-run",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/run",
					Type: newHostPathType("Directory"),
				},
			},
		},
	}

	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ebpfds.Name,
			Namespace: ebpfds.Namespace,
			Labels:    label,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: label,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: label,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "ebpf-agent",
							Image: ebpfds.Spec.Image,
							SecurityContext: &corev1.SecurityContext{
								Privileged: &ebpfds.Spec.RunPrivileged,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "debugfs",
									MountPath: "/sys/kernel/debug",
								},
								{
									Name:      "var-run",
									MountPath: "/var/run",
									ReadOnly:  true,
								},
								{
									Name:      "host-run",
									MountPath: "/run",
									ReadOnly:  true,
								},
							},
							Resources: ebpfds.Spec.Resources,
							Env: []corev1.EnvVar{
								{
									Name:  "TRACER",
									Value: strings.Join(ebpfds.Spec.EnableProbes, ","),
								},
								{
									Name:  "SERVER_ADDR",
									Value: ebpfds.Spec.ServerAddress,
								},
								{
									Name:  "SERVER_PORT",
									Value: ebpfds.Spec.ServerPort,
								},
								{
									Name: "NODE_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
							},
						},
					},
					AutomountServiceAccountToken: ptr.To(false),
					Tolerations:                  ebpfds.Spec.Tolerations,
					NodeSelector:                 ebpfds.Spec.NodeSelector,
					Volumes:                      volumes,
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(ebpfds, ds, r.Scheme); err != nil {
		return nil, err
	}
	return ds, nil
}

func (r *EbpfDaemonSetReconciler) CmpDaemonSets(found, desires *appsv1.DaemonSet, log logr.Logger) bool {
	diff := false
	foundSpec := ebpfv1alpha1.ExtractRelevantSpec(found.Spec.Template.Spec)

	desiredSpec := ebpfv1alpha1.ExtractRelevantSpec(desires.Spec.Template.Spec)

	if !equality.Semantic.DeepEqual(foundSpec.Image, desiredSpec.Image) {
		log.Info("Image differs", "found", foundSpec.Image, "desired", desiredSpec.Image)
		diff = true
		found.Spec.Template.Spec.Containers[0].Image = desiredSpec.Image
	}
	if !equality.Semantic.DeepEqual(foundSpec.NodeSelector, desiredSpec.NodeSelector) {
		log.Info("NodeSelector differs", "found", foundSpec.NodeSelector, "desired", desiredSpec.NodeSelector)
		diff = true
		found.Spec.Template.Spec.NodeSelector = desiredSpec.NodeSelector
	}
	if !equality.Semantic.DeepEqual(foundSpec.Resources, desiredSpec.Resources) {
		log.Info("Resources differ", "found", foundSpec.Resources, "desired", desiredSpec.Resources)
		diff = true
		found.Spec.Template.Spec.Containers[0].Resources = desiredSpec.Resources
	}
	if !equality.Semantic.DeepEqual(foundSpec.RunPrivileged, desiredSpec.RunPrivileged) {
		log.Info("RunPrivileged differs", "found", foundSpec.RunPrivileged, "desired", desiredSpec.RunPrivileged)
		diff = true
		for i, c := range found.Spec.Template.Spec.Containers {
			if c.SecurityContext != nil {
				found.Spec.Template.Spec.Containers[i].SecurityContext.Privileged = &desiredSpec.RunPrivileged
			}
		}
	}
	if !equality.Semantic.DeepEqual(foundSpec.Tolerations, desiredSpec.Tolerations) {
		log.Info("Tolerations differ", "found", foundSpec.Tolerations, "desired", desiredSpec.Tolerations)
		diff = true
		found.Spec.Template.Spec.Tolerations = desiredSpec.Tolerations
	}

	if !equality.Semantic.DeepEqual(foundSpec.EnabledProbes, desiredSpec.EnabledProbes) {
		log.Info("EnabledProbes differ", "found", foundSpec.EnabledProbes, "desired", desiredSpec.EnabledProbes)
		diff = true
		updateEnvVar(&found.Spec.Template.Spec.Containers[0], "TRACER", strings.Join(desiredSpec.EnabledProbes, ","))
	}

	if !equality.Semantic.DeepEqual(foundSpec.ServerAddress, desiredSpec.ServerAddress) {
		log.Info("ServerAddress differs", "found", foundSpec.ServerAddress, "desired", desiredSpec.ServerAddress)
		diff = true
		updateEnvVar(&found.Spec.Template.Spec.Containers[0], "SERVER_ADDR", desiredSpec.ServerAddress)
	}

	if !equality.Semantic.DeepEqual(foundSpec.ServerPort, desiredSpec.ServerPort) {
		log.Info("ServerPort differs", "found", foundSpec.ServerPort, "desired", desiredSpec.ServerPort)
		diff = true
		updateEnvVar(&found.Spec.Template.Spec.Containers[0], "SERVER_PORT", desiredSpec.ServerPort)
	}
	return diff
}
func updateEnvVar(container *corev1.Container, name, value string) {
	for i, env := range container.Env {
		if env.Name == name {
			container.Env[i].Value = value
			return
		}
	}
	container.Env = append(container.Env, corev1.EnvVar{
		Name:  name,
		Value: value,
	})
}
