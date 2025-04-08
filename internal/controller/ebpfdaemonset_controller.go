package controller

import (
	"context"
	"fmt"
	"time"

	ebpfdsv1alpha1 "github.com/ALEYI17/kube-ebpf-monitor/pkg/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type EbpfDaemonSetReconciler struct {
	client.Client
	scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ebpf.monitoring.dev,resources=ebpfdaemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ebpf.monitoring.dev,resources=ebpfdaemonsets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ebpf.monitoring.dev,resources=ebpfdaemonsets/finalizers,verbs=update

func (r *EbpfDaemonSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

  // Manage the CR 
	ebpfDs := &ebpfdsv1alpha1.EbpfDaemonSet{}

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
	if err := r.Get(ctx, req.NamespacedName, found); err != nil && apierrors.IsNotFound(err) {
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

	} else if err != nil{
		log.Error(err, "Failed to get DaemonSet")
		return ctrl.Result{}, err
	}

  

  meta.SetStatusCondition(&ebpfDs.Status.Conditions,
			metav1.Condition{Type: "Available", Status: metav1.ConditionTrue, Reason: "Reconciling", 
      Message: fmt.Sprintf("EbpfDaemonSet created successfully" ) })

  if err := r.Status().Update(ctx, ebpfDs) ; err!=nil{
    log.Error(err, "Failed to update status")                                                  
    return ctrl.Result{}, err
  }

	return ctrl.Result{}, nil
}

func (r *EbpfDaemonSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ebpfdsv1alpha1.EbpfDaemonSet{}).
		Owns(&appsv1.DaemonSet{}).
		Named("ebpfmonitor").
		Complete(r)
}

func (r *EbpfDaemonSetReconciler) DaemonSetForEbpf(ebpfds *ebpfdsv1alpha1.EbpfDaemonSet) (*appsv1.DaemonSet, error) {
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

	if err := ctrl.SetControllerReference(ebpfds, ds, r.scheme); err != nil {
		return nil, err
	}
	return ds, nil
}
