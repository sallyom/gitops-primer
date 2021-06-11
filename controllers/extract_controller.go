/*
Copyright 2021.

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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/operator-framework/operator-lib/status"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	primerv1alpha1 "github.com/cooktheryan/gitops-primer/api/v1alpha1"
)

// ExtractReconciler reconciles a Extract object
type ExtractReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

type ExtractionReconciler struct {
	serviceAccount *corev1.ServiceAccount
	job            *batchv1.Job
	Log            logr.Logger
	Scheme         *runtime.Scheme
	Ctx            context.Context
}

//+kubebuilder:rbac:groups=primer.gitops.io,resources=extracts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=primer.gitops.io,resources=extracts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=primer.gitops.io,resources=extracts/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=*,resources=*,verbs=get;list

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Extract object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *ExtractReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	// Fetch the Extract instance
	instance := &primerv1alpha1.Extract{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Extract resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Extract")
		return ctrl.Result{}, err
	}

	// Check if the Job already exists, if not create a new one
	found := &batchv1.Job{}
	err = r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new job
		job := r.jobForExtract(instance)
		log.Info("Creating a new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
		err = r.Create(ctx, job)
		if err != nil {
			log.Error(err, "Failed to create new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
			return ctrl.Result{}, err
		}
		// Job created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Job")
		return ctrl.Result{}, err
	}

	// Check if the Service Account already exists, if not create a new one
	foundSA := &corev1.ServiceAccount{}
	err = r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, foundSA)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Service Account
		serviceAcct := r.saGenerate(instance)
		log.Info("Creating a new Service Account", "serviceAcct.Namespace", serviceAcct.Namespace, "serviceAcct.Name", serviceAcct.Name)
		err = r.Create(ctx, serviceAcct)
		if err != nil {
			log.Error(err, "Failed to create new Service Account", "serviceAcct.Namespace", serviceAcct.Namespace, "serviceAcct.Name", serviceAcct.Name)
			return ctrl.Result{}, err
		}
		// Service Account created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Service Account")
		return ctrl.Result{}, err
	}

	// Check if the Role already exists, if not create a new one
	foundRole := &rbacv1.Role{}
	err = r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, foundRole)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Role
		role := r.roleGenerate(instance)
		log.Info("Creating a new Role", "role.Namespace", role.Namespace, "role.Name", role.Name)
		err = r.Create(ctx, role)
		if err != nil {
			log.Error(err, "Failed to create new Role", "role.Namespace", role.Namespace, "role.Name", role.Name)
			return ctrl.Result{}, err
		}
		// Role created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Role")
		return ctrl.Result{}, err
	}

	// Check if the RoleBinding already exists, if not create a new one
	foundRoleBinding := &rbacv1.RoleBinding{}
	err = r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, foundRoleBinding)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Role Binding
		roleBinding := r.roleBindingGenerate(instance)
		log.Info("Creating a new Role Binding", "roleBinding.Namespace", roleBinding.Namespace, "roleBinding.Name", roleBinding.Name)
		err = r.Create(ctx, roleBinding)
		if err != nil {
			log.Error(err, "Failed to create new Role Binding", "roleBinding.Namespace", roleBinding.Namespace, "roleBinding.Name", roleBinding.Name)
			return ctrl.Result{}, err
		}
		// Role Binding created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Role Binding")
		return ctrl.Result{}, err
	}

	// Update the Extract status with the pod names
	// List the pods for this instance's Job
	jobState := &batchv1.JobList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
	}
	if err = r.List(ctx, jobState, listOpts...); err != nil {
		log.Error(err, "Failed to list jobs", "Extract.Namespace", instance.Namespace, "Extract.Name", instance.Name)
		return ctrl.Result{}, err
	}

	if instance.Status.Conditions == nil {
		instance.Status.Conditions = status.Conditions{}
	}

	// Set reconcile status condition
	if err == nil {
		instance.Status.Conditions.SetCondition(
			status.Condition{
				Type:    primerv1alpha1.ConditionReconciled,
				Status:  corev1.ConditionTrue,
				Reason:  primerv1alpha1.ReconciledReasonComplete,
				Message: "Reconcile complete",
			})
	} else {
		instance.Status.Conditions.SetCondition(
			status.Condition{
				Type:    primerv1alpha1.ConditionReconciled,
				Status:  corev1.ConditionFalse,
				Reason:  primerv1alpha1.ReconciledReasonError,
				Message: err.Error(),
			})
	}

	statusErr := r.Client.Status().Update(ctx, instance)
	if err == nil { // Don't mask previous error
		err = statusErr
	}

	return ctrl.Result{}, nil
}

// reconcileFunc is a function that partially reconciles an object. It returns a
// bool indicating whether reconciling should continue and an error.
type reconcileFunc func(logr.Logger) (bool, error)

// reconcileBatch steps through a list of reconcile functions until one returns
// false or an error.
func reconcileBatch(l logr.Logger, reconcileFuncs ...reconcileFunc) (bool, error) {
	for _, f := range reconcileFuncs {
		if cont, err := f(l); !cont || err != nil {
			return cont, err
		}
	}
	return true, nil
}

func RunExtactBatch(
	ctx context.Context,
	instance *primerv1alpha1.Extract,
	sr *ExtractReconciler,
	logger logr.Logger,
) (ctrl.Result, error) {
	r := ExtractionReconciler{
		Ctx:                         ctx,
		Instance:                    instance,
		ReplicationSourceReconciler: *sr,
	}
	l := logger.WithValues("Extraction")

	_, err := reconcileBatch(l,
		r.EnsurePVC,
		r.pvcForCache,
		r.ensureServiceAccount,
		r.ensureRepository,
		r.ensureJob,
		r.cleanupJob,
		r.CleanupPVC,
	)
	return ctrl.Result{}, err
}

// jobForExtract returns a instance Job object
func (r *ExtractionReconciler) jobForExtract(m *primerv1alpha1.Extract) *batchv1.Job {
	mode := int32(0600)
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "primer-extract" + m.Name,
			Namespace: m.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy:      "Never",
					ServiceAccountName: m.Name,
					Containers: []corev1.Container{{
						Image:   "quay.io/octo-emerging/gitops-primer-extract:latest",
						Name:    m.Name,
						Command: []string{"/bin/sh", "-c", "/committer.sh"},
						Env: []corev1.EnvVar{
							{Name: "REPO", Value: m.Spec.Repo},
							{Name: "BRANCH", Value: m.Spec.Branch},
							{Name: "ACTION", Value: m.Spec.Action},
						},
						VolumeMounts: []corev1.VolumeMount{
							{Name: "sshkeys", MountPath: "/keys"},
							{Name: "repo", MountPath: "/repo"},
						},
					}},
					Volumes: []corev1.Volume{
						{Name: "repo", VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
						},
						{Name: "sshkeys", VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName:  m.Spec.Secret,
								DefaultMode: &mode,
							}},
						},
					},
				},
			},
		},
	}
	// Set Memcached instance as the owner and controller
	ctrl.SetControllerReference(m, job, r.Scheme)
	return job
}

func (r *ExtractReconciler) saGenerate(m *primerv1alpha1.Extract) *corev1.ServiceAccount {
	// Define a new Service Account object
	serviceAcct := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
	}
	// Service reconcile finished
	ctrl.SetControllerReference(m, serviceAcct, r.Scheme)
	return serviceAcct
}

func (r *ExtractReconciler) roleGenerate(m *primerv1alpha1.Extract) *rbacv1.Role {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"get", "list"},
			},
		},
	}
	// Service reconcile finished
	ctrl.SetControllerReference(m, role, r.Scheme)
	return role
}

func (r *ExtractReconciler) roleBindingGenerate(m *primerv1alpha1.Extract) *rbacv1.RoleBinding {
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Name:     m.Name,
			Kind:     "Role",
		},
		Subjects: []rbacv1.Subject{
			{Kind: "ServiceAccount", Name: m.Name},
		},
	}
	// Service reconcile finished
	ctrl.SetControllerReference(m, roleBinding, r.Scheme)
	return roleBinding
}

func isJobComplete(job *batchv1.Job) bool {
	state := job.Status.Succeeded
	jobComplete := false
	if state == 1 {
		jobComplete = true
	}
	return jobComplete
}

// SetupWithManager sets up the controller with the Manager.
func (r *ExtractReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&primerv1alpha1.Extract{}).
		Owns(&batchv1.Job{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&corev1.ServiceAccount{}).
		Complete(r)
}
