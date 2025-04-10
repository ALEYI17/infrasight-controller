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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	ebpfv1alpha1 "github.com/ALEYI17/kube-ebpf-monitor/api/v1alpha1"
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
		meta.SetStatusCondition(&ebpfDs.Status.Conditions,
			metav1.Condition{Type: "Available", Status: metav1.ConditionUnknown, Reason: "Reconciling", Message: "Starting reconciliation"})

		if err := r.Status().Update(ctx, ebpfDs); err != nil {
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
	if err := r.Get(ctx, types.NamespacedName{Name: ebpfDs.Name,Namespace: ebpfDs.Namespace}, found); err != nil && apierrors.IsNotFound(err) {
		ds, err := r.DaemonSetForEbpf(ebpfDs)
		if err != nil {

			log.Error(err, "Failed to define new DaemonSet resource for ebpf")
			meta.SetStatusCondition(&ebpfDs.Status.Conditions, metav1.Condition{Type: "Available", Status: metav1.ConditionFalse, Reason: "Reconciling", Message: fmt.Sprintf("Failed to create DaemonSet for cr (%s): (%s) ", ebpfDs.Name, err)})

			if err := r.Status().Update(ctx, ebpfDs); err != nil {
				log.Error(err, "Failed to update Status")
				return ctrl.Result{}, err
			}
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

	if !equality.Semantic.DeepEqual(desired.Spec.Template, found.Spec.Template) {
		log.Info("Spec change detected, updating daemonset")
		found.Spec.Template = desired.Spec.Template

		err := r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Error updating daemonset")
			return ctrl.Result{}, err
		}

    latest := &ebpfv1alpha1.EbpfDaemonSet{}
    if err := r.Get(ctx, req.NamespacedName, latest); err != nil {
      log.Error(err, "Failed to re-fetch EbpfDaemonSet before updating status")
      return ctrl.Result{}, err
    }

		meta.SetStatusCondition(&latest.Status.Conditions,
			metav1.Condition{Type: "Available", Status: metav1.ConditionUnknown, Reason: "Reconciling",
				Message: fmt.Sprintf("EbpfDaemonSet is updating")})
		if err := r.Status().Update(ctx, latest); err != nil {
			log.Error(err, "Failed to update status")
			return ctrl.Result{}, err

		}

		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	meta.SetStatusCondition(&ebpfDs.Status.Conditions,
		metav1.Condition{Type: "Available", Status: metav1.ConditionTrue, Reason: "Reconciling",
			Message: fmt.Sprintf("EbpfDaemonSet created successfully")})

	if err := r.Status().Update(ctx, ebpfDs); err != nil {
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

func (r *EbpfDaemonSetReconciler) DaemonSetForEbpf(ebpfds *ebpfv1alpha1.EbpfDaemonSet) (*appsv1.DaemonSet, error) {
	label := map[string]string{
		"app": ebpfds.Name,
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
							Resources: ebpfds.Spec.Resources,
						},
					},
					Tolerations:  ebpfds.Spec.Tolerations,
					NodeSelector: ebpfds.Spec.NodeSelector,
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(ebpfds, ds, r.Scheme); err != nil {
		return nil, err
	}
	return ds, nil
}
