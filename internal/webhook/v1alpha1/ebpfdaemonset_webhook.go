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
	"strconv"
	"strings"

	ebpfv1alpha1 "github.com/ALEYI17/kube-ebpf-monitor/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
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

// +kubebuilder:webhook:path=/mutate-ebpf-monitoring-dev-v1alpha1-ebpfdaemonset,mutating=true,failurePolicy=fail,sideEffects=None,groups=ebpf.monitoring.dev,resources=ebpfdaemonsets,verbs=create;update,versions=v1alpha1,name=mebpfdaemonset-v1alpha1.kb.io,admissionReviewVersions=v1

// EbpfDaemonSetCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind EbpfDaemonSet when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type EbpfDaemonSetCustomDefaulter struct {
}

var _ webhook.CustomDefaulter = &EbpfDaemonSetCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind EbpfDaemonSet.
func (d *EbpfDaemonSetCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	ebpfdaemonset, ok := obj.(*ebpfv1alpha1.EbpfDaemonSet)

	if !ok {
		return fmt.Errorf("expected an EbpfDaemonSet object but got %T", obj)
	}
	ebpfdaemonsetlog.Info("Defaulting for EbpfDaemonSet", "name", ebpfdaemonset.GetName())

	if ebpfdaemonset.Spec.Image == "" {
		ebpfdaemonset.Spec.Image = "alejandrosalamanca17/knative-example-native:latest"
	}

	if ebpfdaemonset.Spec.NodeSelector == nil {
		ebpfdaemonset.Spec.NodeSelector = map[string]string{"kubernetes.io/os": "linux"}
	}

	if ebpfdaemonset.Spec.Resources.Limits == nil && ebpfdaemonset.Spec.Resources.Requests == nil {
		ebpfdaemonset.Spec.Resources = corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("128Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("512Mi"),
			},
		}
	}

	if !ebpfdaemonset.Spec.RunPrivileged {
		ebpfdaemonset.Spec.RunPrivileged = true
	}
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
}

var _ webhook.CustomValidator = &EbpfDaemonSetCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type EbpfDaemonSet.
func (v *EbpfDaemonSetCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ebpfdaemonset, ok := obj.(*ebpfv1alpha1.EbpfDaemonSet)
	if !ok {
		return nil, fmt.Errorf("expected a EbpfDaemonSet object but got %T", obj)
	}
	ebpfdaemonsetlog.Info("Validation for EbpfDaemonSet upon creation", "name", ebpfdaemonset.GetName())

	return nil, validateEbpfDaemonset(ebpfdaemonset)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type EbpfDaemonSet.
func (v *EbpfDaemonSetCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ebpfdaemonset, ok := newObj.(*ebpfv1alpha1.EbpfDaemonSet)
	if !ok {
		return nil, fmt.Errorf("expected a EbpfDaemonSet object for the newObj but got %T", newObj)
	}
	ebpfdaemonsetlog.Info("Validation for EbpfDaemonSet upon update", "name", ebpfdaemonset.GetName())

	return nil, validateEbpfDaemonset(ebpfdaemonset)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type EbpfDaemonSet.
func (v *EbpfDaemonSetCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ebpfdaemonset, ok := obj.(*ebpfv1alpha1.EbpfDaemonSet)
	if !ok {
		return nil, fmt.Errorf("expected a EbpfDaemonSet object but got %T", obj)
	}
	ebpfdaemonsetlog.Info("Validation for EbpfDaemonSet upon deletion", "name", ebpfdaemonset.GetName())

	return nil, nil
}

func validateEbpfDaemonset(ebpfds *ebpfv1alpha1.EbpfDaemonSet) error {
	var allErrs field.ErrorList

	if err := validateImage(ebpfds); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := validateResources(ebpfds); err != nil {
		allErrs = append(allErrs, *err...)
	}

	if err := validateNodeSelector(ebpfds); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := validateEnableProbes(ebpfds); err != nil {
		allErrs = append(allErrs, *err...)
	}

	if err := ValidateServerAddr(ebpfds); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := ValidateServerPort(ebpfds); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{
		Group: "ebpf.monitoring.dev",
		Kind:  "EbpfDaemonSet",
	},
		ebpfds.Name, allErrs)
}

func validateImage(ebpfds *ebpfv1alpha1.EbpfDaemonSet) *field.Error {
	if ebpfds.Spec.Image == "" {
		return field.Required(field.NewPath("spec").Child("image"), "Image not provide")
	}

	return nil
}

func validateResources(ebpfds *ebpfv1alpha1.EbpfDaemonSet) *field.ErrorList {
	var errs field.ErrorList

	resPath := field.NewPath("spec").Child("resources").Child("requests")
	if cpuReq, ok := ebpfds.Spec.Resources.Requests[corev1.ResourceCPU]; ok {
		if cpuLim, ok := ebpfds.Spec.Resources.Limits[corev1.ResourceCPU]; ok && cpuReq.Cmp(cpuLim) > 0 {
			errs = append(errs, field.Forbidden(resPath.Key(string(corev1.ResourceCPU)), "Cpu Request > Cpu Limits"))
		}
	}

	if memReq, ok := ebpfds.Spec.Resources.Requests[corev1.ResourceMemory]; ok {
		if memLim, ok := ebpfds.Spec.Resources.Limits[corev1.ResourceMemory]; ok && memReq.Cmp(memLim) > 0 {
			errs = append(errs, field.Forbidden(resPath.Key(string(corev1.ResourceMemory)), "Memory request cannot exceed memory limit"))
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return &errs
}

func validateNodeSelector(ebpfds *ebpfv1alpha1.EbpfDaemonSet) *field.Error {

	for key, value := range ebpfds.Spec.NodeSelector {

		if key == "kubernetes.io/os" && value != "linux" {
			return field.Forbidden(field.NewPath("spec").Child("nodeSelector"), "Ebpf only workls on linux nodes")
		}
	}

	return nil
}

func validateEnableProbes(ebpfds *ebpfv1alpha1.EbpfDaemonSet) *field.ErrorList {
	var errs field.ErrorList

	resPath := field.NewPath("spec").Child("EnableProbes")
	if len(ebpfds.Spec.EnableProbes) == 0 {
		errs = append(errs, field.Required(resPath, "at least one probe must be specified"))
	}

	allowed := sets.New[string]("open", "execve","chmod","connect","accept")

	seen := sets.New[string]()

	for i, probe := range ebpfds.Spec.EnableProbes {
		if !allowed.Has(probe) {
			errs = append(errs, field.NotSupported(resPath.Index(i), probe, sets.List(allowed)))
		}

		if seen.Has(probe) {
			errs = append(errs, field.Duplicate(resPath.Index(i), probe))
		} else {
			seen.Insert(probe)
		}

		if probe != strings.ToLower(probe) {
			errs = append(errs, field.Invalid(resPath.Index(i), probe, "probes must be lowercase"))
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return &errs
}

func ValidateServerAddr(ebpfds *ebpfv1alpha1.EbpfDaemonSet) *field.Error {

	if ebpfds.Spec.ServerAddress == "" {
		return field.Required(field.NewPath("spec").Child("serverAddress"), "server address should be provided")
	}

	return nil
}

func ValidateServerPort(ebpfds *ebpfv1alpha1.EbpfDaemonSet) *field.Error {

	if ebpfds.Spec.ServerPort == "" {
		return field.Required(field.NewPath("spec").Child("serverPort"), "server port should be provided")
	}

	port, err := strconv.Atoi(ebpfds.Spec.ServerPort)
	if err != nil || port < 1 || port > 65535 {
		return field.Invalid(field.NewPath("spec").Child("serverPort"), ebpfds.Spec.ServerPort, "must be a valid port number (1–65535)")
	}
	return nil
}
