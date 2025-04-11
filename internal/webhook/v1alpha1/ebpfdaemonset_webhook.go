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

package v1alpha1

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	ebpfv1alpha1 "github.com/ALEYI17/kube-ebpf-monitor/api/v1alpha1"
)

// nolint:unused
// log is for logging in this package.
var ebpfdaemonsetlog = logf.Log.WithName("ebpfdaemonset-resource")

// SetupEbpfDaemonSetWebhookWithManager registers the webhook for EbpfDaemonSet in the manager.
func SetupEbpfDaemonSetWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&ebpfv1alpha1.EbpfDaemonSet{}).
		WithValidator(&EbpfDaemonSetCustomValidator{}).
		WithDefaulter(&EbpfDaemonSetCustomDefaulter{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-ebpf-monitoring-dev-v1alpha1-ebpfdaemonset,mutating=true,failurePolicy=fail,sideEffects=None,groups=ebpf.monitoring.dev,resources=ebpfdaemonsets,verbs=create;update,versions=v1alpha1,name=mebpfdaemonset-v1alpha1.kb.io,admissionReviewVersions=v1

// EbpfDaemonSetCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind EbpfDaemonSet when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type EbpfDaemonSetCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
}

var _ webhook.CustomDefaulter = &EbpfDaemonSetCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind EbpfDaemonSet.
func (d *EbpfDaemonSetCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	ebpfdaemonset, ok := obj.(*ebpfv1alpha1.EbpfDaemonSet)

	if !ok {
		return fmt.Errorf("expected an EbpfDaemonSet object but got %T", obj)
	}
	ebpfdaemonsetlog.Info("Defaulting for EbpfDaemonSet", "name", ebpfdaemonset.GetName())

	// TODO(user): fill in your defaulting logic.

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-ebpf-monitoring-dev-v1alpha1-ebpfdaemonset,mutating=false,failurePolicy=fail,sideEffects=None,groups=ebpf.monitoring.dev,resources=ebpfdaemonsets,verbs=create;update,versions=v1alpha1,name=vebpfdaemonset-v1alpha1.kb.io,admissionReviewVersions=v1

// EbpfDaemonSetCustomValidator struct is responsible for validating the EbpfDaemonSet resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type EbpfDaemonSetCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &EbpfDaemonSetCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type EbpfDaemonSet.
func (v *EbpfDaemonSetCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ebpfdaemonset, ok := obj.(*ebpfv1alpha1.EbpfDaemonSet)
	if !ok {
		return nil, fmt.Errorf("expected a EbpfDaemonSet object but got %T", obj)
	}
	ebpfdaemonsetlog.Info("Validation for EbpfDaemonSet upon creation", "name", ebpfdaemonset.GetName())

	// TODO(user): fill in your validation logic upon object creation.

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type EbpfDaemonSet.
func (v *EbpfDaemonSetCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ebpfdaemonset, ok := newObj.(*ebpfv1alpha1.EbpfDaemonSet)
	if !ok {
		return nil, fmt.Errorf("expected a EbpfDaemonSet object for the newObj but got %T", newObj)
	}
	ebpfdaemonsetlog.Info("Validation for EbpfDaemonSet upon update", "name", ebpfdaemonset.GetName())

	// TODO(user): fill in your validation logic upon object update.

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type EbpfDaemonSet.
func (v *EbpfDaemonSetCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ebpfdaemonset, ok := obj.(*ebpfv1alpha1.EbpfDaemonSet)
	if !ok {
		return nil, fmt.Errorf("expected a EbpfDaemonSet object but got %T", obj)
	}
	ebpfdaemonsetlog.Info("Validation for EbpfDaemonSet upon deletion", "name", ebpfdaemonset.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}
