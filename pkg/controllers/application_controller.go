// pkg/controllers/application_controller.go
// The brain of our platform - watches for Application resources and makes things happen!

package controllers

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/virtual457/orion-platform/pkg/apis/platform/v1alpha1"
)

// ApplicationController manages the lifecycle of Application resources
type ApplicationController struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile is the main controller logic - this gets called whenever an Application changes
func (r *ApplicationController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("🔄 Reconciling Application", "name", req.Name, "namespace", req.Namespace)

	// Fetch the Application resource that triggered this reconciliation
	app := &v1alpha1.Application{}
	err := r.Get(ctx, req.NamespacedName, app)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("❌ Application not found - might have been deleted", "name", req.Name)
			return ctrl.Result{}, nil
		}
		logger.Error(err, "❌ Failed to get Application")
		return ctrl.Result{}, err
	}

	logger.Info("📋 Found Application", "image", app.Spec.Image, "replicas", app.GetReplicas())

	// Validate the Application spec
	if err := app.ValidateSpec(); err != nil {
		logger.Error(err, "❌ Application spec validation failed")
		app.UpdateStatus(v1alpha1.PhaseFailed, fmt.Sprintf("Validation failed: %v", err))
		return r.updateApplicationStatus(ctx, app)
	}

	// Main reconciliation logic
	return r.reconcileApplication(ctx, app)
}

// reconcileApplication handles the main application lifecycle
func (r *ApplicationController) reconcileApplication(ctx context.Context, app *v1alpha1.Application) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	
	// Phase 1: Provision Infrastructure (if needed)
	if app.Status.Phase == "" || app.Status.Phase == v1alpha1.PhasePending {
		logger.Info("🏗️ Starting infrastructure provisioning")
		app.UpdateStatus(v1alpha1.PhaseProvisioningInfra, "Provisioning AWS infrastructure")
		
		if err := r.updateApplicationStatusOnly(ctx, app); err != nil {
			return ctrl.Result{}, err
		}
		
		// Simulate infrastructure provisioning
		if err := r.provisionInfrastructure(ctx, app); err != nil {
			logger.Error(err, "❌ Infrastructure provisioning failed")
			app.UpdateStatus(v1alpha1.PhaseFailed, fmt.Sprintf("Infrastructure failed: %v", err))
			r.updateApplicationStatusOnly(ctx, app)
			return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
		}
		
		// Requeue to continue with deployment
		return ctrl.Result{RequeueAfter: time.Second * 10}, nil
	}

	// Phase 2: Deploy Application
	if app.Status.Phase == v1alpha1.PhaseProvisioningInfra && app.Status.InfrastructureReady {
		logger.Info("🚀 Starting application deployment")
		app.UpdateStatus(v1alpha1.PhaseDeploying, "Creating Kubernetes resources")
		
		if err := r.updateApplicationStatusOnly(ctx, app); err != nil {
			return ctrl.Result{}, err
		}
		
		// Create Kubernetes Deployment
		if err := r.createOrUpdateDeployment(ctx, app); err != nil {
			logger.Error(err, "❌ Failed to create deployment")
			app.UpdateStatus(v1alpha1.PhaseFailed, fmt.Sprintf("Deployment failed: %v", err))
			r.updateApplicationStatusOnly(ctx, app)
			return ctrl.Result{RequeueAfter: time.Minute * 2}, nil
		}

		// Create Kubernetes Service
		if err := r.createOrUpdateService(ctx, app); err != nil {
			logger.Error(err, "❌ Failed to create service")
			app.UpdateStatus(v1alpha1.PhaseFailed, fmt.Sprintf("Service failed: %v", err))
			r.updateApplicationStatusOnly(ctx, app)
			return ctrl.Result{RequeueAfter: time.Minute * 2}, nil
		}

		// Requeue to check if deployment is ready
		return ctrl.Result{RequeueAfter: time.Second * 15}, nil
	}

	// Phase 3: Check if Application is Ready
	if app.Status.Phase == v1alpha1.PhaseDeploying {
		ready, err := r.checkApplicationReady(ctx, app)
		if err != nil {
			logger.Error(err, "❌ Failed to check application readiness")
			return ctrl.Result{RequeueAfter: time.Second * 30}, nil
		}

		if ready {
			logger.Info("✅ Application is ready!")
			app.UpdateStatus(v1alpha1.PhaseReady, "All replicas ready and serving traffic")
			return r.updateApplicationStatus(ctx, app)
		}

		// Still deploying, check again later
		logger.Info("⏳ Application still deploying...")
		return ctrl.Result{RequeueAfter: time.Second * 15}, nil
	}

	// Application is ready - periodic health check
	if app.Status.Phase == v1alpha1.PhaseReady {
		logger.Info("💚 Application healthy - periodic check")
		return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
	}

	logger.Info("🤔 Unknown phase", "phase", app.Status.Phase)
	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

// provisionInfrastructure handles AWS resource provisioning
func (r *ApplicationController) provisionInfrastructure(ctx context.Context, app *v1alpha1.Application) error {
	logger := log.FromContext(ctx)
	
	// Simulate infrastructure provisioning
	// In production, this would call AWS APIs or Terraform
	
	if app.NeedsDatabase() {
		logger.Info("📊 Provisioning PostgreSQL database")
		// TODO: Call AWS RDS API
		app.Status.DatabaseEndpoint = fmt.Sprintf("%s-db.cluster-xyz.us-west-2.rds.amazonaws.com", app.Name)
	}
	
	if app.NeedsCache() {
		logger.Info("🔄 Provisioning Redis cache")
		// TODO: Call AWS ElastiCache API
		app.Status.RedisEndpoint = fmt.Sprintf("%s-cache.xyz.cache.amazonaws.com", app.Name)
	}
	
	if app.NeedsStorage() {
		logger.Info("🪣 Provisioning S3 bucket")
		// TODO: Call AWS S3 API
		app.Status.S3BucketName = fmt.Sprintf("%s-storage-%d", app.Name, time.Now().Unix())
	}
	
	// Mark infrastructure as ready
	app.Status.InfrastructureReady = true
	logger.Info("✅ Infrastructure provisioning complete")
	
	return nil
}

// createOrUpdateDeployment creates a Kubernetes Deployment for the application
func (r *ApplicationController) createOrUpdateDeployment(ctx context.Context, app *v1alpha1.Application) error {
	logger := log.FromContext(ctx)
	
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			Labels:    map[string]string{"app": app.Name, "managed-by": "orion-platform"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &[]int32{app.GetReplicas()}[0],
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": app.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": app.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  app.Name,
							Image: app.Spec.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: app.GetPort(),
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env: r.buildEnvironmentVariables(app),
						},
					},
				},
			},
		},
	}

	// Try to create the deployment
	if err := r.Create(ctx, deployment); err != nil {
		if errors.IsAlreadyExists(err) {
			logger.Info("📦 Deployment already exists, updating...")
			// TODO: Update existing deployment
			return nil
		}
		return fmt.Errorf("failed to create deployment: %w", err)
	}

	logger.Info("✅ Created Kubernetes Deployment", "replicas", app.GetReplicas())
	return nil
}

// createOrUpdateService creates a Kubernetes Service for the application
func (r *ApplicationController) createOrUpdateService(ctx context.Context, app *v1alpha1.Application) error {
	logger := log.FromContext(ctx)
	
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			Labels:    map[string]string{"app": app.Name, "managed-by": "orion-platform"},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": app.Name},
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt32(app.GetPort()),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	if err := r.Create(ctx, service); err != nil {
		if errors.IsAlreadyExists(err) {
			logger.Info("🌐 Service already exists")
			return nil
		}
		return fmt.Errorf("failed to create service: %w", err)
	}

	logger.Info("✅ Created Kubernetes Service", "port", app.GetPort())
	return nil
}

// buildEnvironmentVariables creates environment variables including infrastructure credentials
func (r *ApplicationController) buildEnvironmentVariables(app *v1alpha1.Application) []corev1.EnvVar {
	envVars := []corev1.EnvVar{}

	// Add user-defined environment variables
	for key, value := range app.Spec.Env {
		envVars = append(envVars, corev1.EnvVar{Name: key, Value: value})
	}

	// Add infrastructure connection details
	if app.Status.DatabaseEndpoint != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "DATABASE_URL",
			Value: fmt.Sprintf("postgres://user:password@%s:5432/%s", app.Status.DatabaseEndpoint, "app"),
		})
	}

	if app.Status.RedisEndpoint != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "REDIS_URL",
			Value: fmt.Sprintf("redis://%s:6379", app.Status.RedisEndpoint),
		})
	}

	if app.Status.S3BucketName != "" {
		envVars = append(envVars, corev1.EnvVar{Name: "S3_BUCKET", Value: app.Status.S3BucketName})
	}

	return envVars
}

// checkApplicationReady checks if the deployment is ready
func (r *ApplicationController) checkApplicationReady(ctx context.Context, app *v1alpha1.Application) (bool, error) {
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, client.ObjectKey{Name: app.Name, Namespace: app.Namespace}, deployment)
	if err != nil {
		return false, err
	}

	// Check if all replicas are ready
	if deployment.Status.ReadyReplicas == app.GetReplicas() {
		app.Status.ReadyReplicas = deployment.Status.ReadyReplicas
		return true, nil
	}

	app.Status.ReadyReplicas = deployment.Status.ReadyReplicas
	return false, nil
}

// updateApplicationStatus updates the Application status in Kubernetes (returns ctrl.Result, error)
func (r *ApplicationController) updateApplicationStatus(ctx context.Context, app *v1alpha1.Application) (ctrl.Result, error) {
	if err := r.Status().Update(ctx, app); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update Application status: %w", err)
	}
	return ctrl.Result{}, nil
}

// updateApplicationStatusOnly updates the Application status (returns error only)
func (r *ApplicationController) updateApplicationStatusOnly(ctx context.Context, app *v1alpha1.Application) error {
	if err := r.Status().Update(ctx, app); err != nil {
		return fmt.Errorf("failed to update Application status: %w", err)
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager
func (r *ApplicationController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Application{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}