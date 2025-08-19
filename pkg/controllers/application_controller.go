// pkg/controllers/application_controller.go
// Enhanced with Smart Environment Selection

package controllers

import (
	"context"
	"fmt"
	"os"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
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

// Reconcile is the main controller logic - enhanced with environment awareness
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

	logger.Info("📋 Found Application", 
		"image", app.Spec.Image, 
		"replicas", app.GetReplicas(),
		"infrastructure", app.GetInfrastructureSummary())

	// Validate the Application spec
	if err := app.ValidateSpec(); err != nil {
		logger.Error(err, "❌ Application spec validation failed")
		app.UpdateStatus(v1alpha1.PhaseFailed, fmt.Sprintf("Validation failed: %v", err))
		return r.updateApplicationStatus(ctx, app)
	}

	// Main reconciliation logic
	return r.reconcileApplication(ctx, app)
}

// reconcileApplication handles the main application lifecycle with environment awareness
func (r *ApplicationController) reconcileApplication(ctx context.Context, app *v1alpha1.Application) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	
	// Phase 1: Provision Infrastructure (environment-aware)
	if app.Status.Phase == "" || app.Status.Phase == v1alpha1.PhasePending {
		logger.Info("🏗️ Starting environment-aware infrastructure provisioning")
		app.UpdateStatus(v1alpha1.PhaseProvisioningInfra, "Analyzing environment and provisioning infrastructure")
		
		if err := r.updateApplicationStatusOnly(ctx, app); err != nil {
			return ctrl.Result{}, err
		}
		
		// Smart infrastructure provisioning
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

// provisionInfrastructure handles environment-aware resource provisioning
func (r *ApplicationController) provisionInfrastructure(ctx context.Context, app *v1alpha1.Application) error {
	logger := log.FromContext(ctx)
	
	// Provision PostgreSQL
	if app.NeedsDatabase() {
		if app.IsLocalDatabase() {
			logger.Info("🏠 Provisioning local PostgreSQL")
			if err := r.provisionLocalPostgreSQL(ctx, app); err != nil {
				return fmt.Errorf("failed to provision local PostgreSQL: %w", err)
			}
			logger.Info("✅ Local PostgreSQL provisioned", "endpoint", app.Status.DatabaseEndpoint)
		} else {
			if err := r.provisionAWSPostgreSQL(ctx, app); err != nil {
				return fmt.Errorf("failed to provision AWS PostgreSQL: %w", err)
			}
		}
	}
	
	// Provision Redis
	if app.NeedsCache() {
		if app.IsLocalRedis() {
			logger.Info("🏠 Provisioning local Redis")
			if err := r.provisionLocalRedis(ctx, app); err != nil {
				return fmt.Errorf("failed to provision local Redis: %w", err)
			}
			logger.Info("✅ Local Redis provisioned", "endpoint", app.Status.RedisEndpoint)
		} else {
			if err := r.provisionAWSRedis(ctx, app); err != nil {
				return fmt.Errorf("failed to provision AWS Redis: %w", err)
			}
		}
	}
	
	// Provision S3/Storage
	if app.NeedsStorage() {
		if app.IsLocalS3() {
			logger.Info("🏠 Provisioning local S3 (MinIO)")
			if err := r.provisionLocalS3(ctx, app); err != nil {
				return fmt.Errorf("failed to provision local S3 (MinIO): %w", err)
			}
			logger.Info("✅ Local S3 provisioned", "endpoint", app.Status.S3Endpoint)
		} else {
			if err := r.provisionAWSS3(ctx, app); err != nil {
				return fmt.Errorf("failed to provision AWS S3: %w", err)
			}
		}
	}
	
	// CRITICAL: Mark infrastructure as ready and update status immediately
	app.Status.InfrastructureReady = true
	logger.Info("✅ All infrastructure provisioned - updating status")
	
	// Update status in Kubernetes
	if err := r.Status().Update(ctx, app); err != nil {
		logger.Error(err, "Failed to update infrastructure status")
		return fmt.Errorf("failed to update infrastructure status: %w", err)
	}
	
	logger.Info("🎉 Infrastructure provisioning complete and status updated")
	return nil
}

// provisionLocalPostgreSQL creates a local PostgreSQL with persistent storage
func (r *ApplicationController) provisionLocalPostgreSQL(ctx context.Context, app *v1alpha1.Application) error {
	logger := log.FromContext(ctx)
	logger.Info("🏠 Creating local PostgreSQL with persistent storage")
	
	// Step 1: Create Persistent Volume Claim
	storageSize := "2Gi" // Default
	if app.Spec.Infrastructure.PostgreSQL.LocalStorage != "" {
		storageSize = app.Spec.Infrastructure.PostgreSQL.LocalStorage
	}
	
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-postgres-pvc", app.Name),
			Namespace: app.Namespace,
			Labels:    map[string]string{"app": app.Name, "component": "database", "managed-by": "orion-platform"},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(storageSize),
				},
			},
		},
	}
	
	if err := r.Create(ctx, pvc); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create PostgreSQL PVC: %w", err)
	}
	
	// Step 2: Create StatefulSet with persistent storage
	dbName := "webapp"
	if app.Spec.Infrastructure.PostgreSQL.DatabaseName != "" {
		dbName = app.Spec.Infrastructure.PostgreSQL.DatabaseName
	}
	
	postgres := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-postgres", app.Name),
			Namespace: app.Namespace,
			Labels:    map[string]string{"app": app.Name, "component": "database", "managed-by": "orion-platform"},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &[]int32{1}[0],
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": app.Name, "component": "database"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": app.Name, "component": "database"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "postgres",
							Image: fmt.Sprintf("postgres:%s", app.Spec.Infrastructure.PostgreSQL.Version),
							Env: []corev1.EnvVar{
								{Name: "POSTGRES_DB", Value: dbName},
								{Name: "POSTGRES_USER", Value: "appuser"},
								{Name: "POSTGRES_PASSWORD", Value: "localpassword"},
								{Name: "PGDATA", Value: "/var/lib/postgresql/data/pgdata"},
							},
							Ports: []corev1.ContainerPort{{ContainerPort: 5432}},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "postgres-data",
									MountPath: "/var/lib/postgresql/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "postgres-data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: fmt.Sprintf("%s-postgres-pvc", app.Name),
								},
							},
						},
					},
				},
			},
		},
	}
	
	if err := r.Create(ctx, postgres); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create PostgreSQL StatefulSet: %w", err)
	}
	
	// Step 3: Create Service for database access
	dbService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-postgres", app.Name),
			Namespace: app.Namespace,
			Labels:    map[string]string{"app": app.Name, "component": "database", "managed-by": "orion-platform"},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": app.Name, "component": "database"},
			Ports: []corev1.ServicePort{
				{
					Port:       5432,
					TargetPort: intstr.FromInt(5432),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
	
	if err := r.Create(ctx, dbService); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create PostgreSQL Service: %w", err)
	}
	
	// Update application status
	app.Status.DatabaseEndpoint = fmt.Sprintf("%s-postgres:5432", app.Name)
	app.Status.DatabaseEnvironment = v1alpha1.EnvironmentLocal
	
	logger.Info("✅ Local PostgreSQL created", 
		"endpoint", app.Status.DatabaseEndpoint,
		"storage", storageSize,
		"database", dbName)
	
	return nil
}

// provisionLocalRedis creates a local Redis instance
func (r *ApplicationController) provisionLocalRedis(ctx context.Context, app *v1alpha1.Application) error {
	logger := log.FromContext(ctx)
	logger.Info("🏠 Creating local Redis")
	
	redis := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-redis", app.Name),
			Namespace: app.Namespace,
			Labels:    map[string]string{"app": app.Name, "component": "cache", "managed-by": "orion-platform"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &[]int32{1}[0],
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": app.Name, "component": "cache"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": app.Name, "component": "cache"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "redis",
							Image: fmt.Sprintf("redis:%s", app.Spec.Infrastructure.Redis.Version),
							Ports: []corev1.ContainerPort{{ContainerPort: 6379}},
						},
					},
				},
			},
		},
	}
	
	if err := r.Create(ctx, redis); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create Redis Deployment: %w", err)
	}
	
	// Create Redis Service
	redisService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-redis", app.Name),
			Namespace: app.Namespace,
			Labels:    map[string]string{"app": app.Name, "component": "cache", "managed-by": "orion-platform"},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": app.Name, "component": "cache"},
			Ports: []corev1.ServicePort{
				{
					Port:       6379,
					TargetPort: intstr.FromInt(6379),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
	
	if err := r.Create(ctx, redisService); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create Redis Service: %w", err)
	}
	
	// Update application status
	app.Status.RedisEndpoint = fmt.Sprintf("%s-redis:6379", app.Name)
	app.Status.RedisEnvironment = v1alpha1.EnvironmentLocal
	
	logger.Info("✅ Local Redis created", "endpoint", app.Status.RedisEndpoint)
	return nil
}

// provisionLocalS3 creates a local MinIO (S3-compatible) instance
func (r *ApplicationController) provisionLocalS3(ctx context.Context, app *v1alpha1.Application) error {
	logger := log.FromContext(ctx)
	logger.Info("🏠 Creating local S3 (MinIO)")
	
	minio := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-s3", app.Name),
			Namespace: app.Namespace,
			Labels:    map[string]string{"app": app.Name, "component": "storage", "managed-by": "orion-platform"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &[]int32{1}[0],
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": app.Name, "component": "storage"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": app.Name, "component": "storage"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "minio",
							Image:   "minio/minio:latest",
							Command: []string{"/usr/bin/docker-entrypoint.sh"},
							Args:    []string{"server", "/data", "--console-address", ":9001"},
							Env: []corev1.EnvVar{
								{Name: "MINIO_ROOT_USER", Value: "minioadmin"},
								{Name: "MINIO_ROOT_PASSWORD", Value: "minioadmin"},
							},
							Ports: []corev1.ContainerPort{
								{ContainerPort: 9000}, // API
								{ContainerPort: 9001}, // Console
							},
						},
					},
				},
			},
		},
	}
	
	if err := r.Create(ctx, minio); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create MinIO Deployment: %w", err)
	}
	
	// Create MinIO Service
	minioService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-s3", app.Name),
			Namespace: app.Namespace,
			Labels:    map[string]string{"app": app.Name, "component": "storage", "managed-by": "orion-platform"},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": app.Name, "component": "storage"},
			Ports: []corev1.ServicePort{
				{
					Name:       "api",
					Port:       9000,
					TargetPort: intstr.FromInt(9000),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "console",
					Port:       9001,
					TargetPort: intstr.FromInt(9001),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
	
	if err := r.Create(ctx, minioService); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create MinIO Service: %w", err)
	}
	
	// Update application status
	bucketName := "default-bucket"
	if app.Spec.Infrastructure.S3.BucketName != "" {
		bucketName = app.Spec.Infrastructure.S3.BucketName
	}
	
	app.Status.S3BucketName = bucketName
	app.Status.S3Endpoint = fmt.Sprintf("%s-s3:9000", app.Name)
	app.Status.S3Environment = v1alpha1.EnvironmentLocal
	
	logger.Info("✅ Local S3 (MinIO) created", 
		"endpoint", app.Status.S3Endpoint,
		"bucket", bucketName,
		"console", fmt.Sprintf("%s-s3:9001", app.Name))
	
	return nil
}

// AWS provisioning methods (simulated for now)
func (r *ApplicationController) provisionAWSPostgreSQL(ctx context.Context, app *v1alpha1.Application) error {
	logger := log.FromContext(ctx)
	logger.Info("☁️ Simulating AWS RDS PostgreSQL provisioning")
	
	// TODO: Real AWS RDS API calls
	app.Status.DatabaseEndpoint = fmt.Sprintf("%s-db.cluster-xyz.us-west-2.rds.amazonaws.com", app.Name)
	app.Status.DatabaseEnvironment = v1alpha1.EnvironmentAWS
	
	logger.Info("✅ AWS RDS PostgreSQL simulated", "endpoint", app.Status.DatabaseEndpoint)
	return nil
}

func (r *ApplicationController) provisionAWSRedis(ctx context.Context, app *v1alpha1.Application) error {
	logger := log.FromContext(ctx)
	logger.Info("☁️ Simulating AWS ElastiCache Redis provisioning")
	
	// TODO: Real AWS ElastiCache API calls
	app.Status.RedisEndpoint = fmt.Sprintf("%s-cache.xyz.cache.amazonaws.com", app.Name)
	app.Status.RedisEnvironment = v1alpha1.EnvironmentAWS
	
	logger.Info("✅ AWS ElastiCache Redis simulated", "endpoint", app.Status.RedisEndpoint)
	return nil
}

func (r *ApplicationController) provisionAWSS3(ctx context.Context, app *v1alpha1.Application) error {
	logger := log.FromContext(ctx)
	logger.Info("☁️ Simulating AWS S3 provisioning")
	
	// TODO: Real AWS S3 API calls
	bucketName := fmt.Sprintf("%s-storage-%d", app.Name, time.Now().Unix())
	if app.Spec.Infrastructure.S3.BucketName != "" {
		bucketName = app.Spec.Infrastructure.S3.BucketName
	}
	
	app.Status.S3BucketName = bucketName
	app.Status.S3Environment = v1alpha1.EnvironmentAWS
	
	logger.Info("✅ AWS S3 simulated", "bucket", bucketName)
	return nil
}

// Environment detection helper
func (r *ApplicationController) isLocalEnvironment() bool {
	// Check for AWS credentials
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") != "" {
		return false
	}
	
	// Check for cloud metadata (simplified)
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		// Check if it's a cloud provider
		if os.Getenv("AWS_REGION") != "" || os.Getenv("GCP_PROJECT") != "" {
			return false
		}
	}
	
	// Default to local
	return true
}

// Existing methods continue below (createOrUpdateDeployment, buildEnvironmentVariables, etc.)
// ... (keeping all the existing methods from the previous controller)

// Enhanced buildEnvironmentVariables with environment-aware connections
func (r *ApplicationController) buildEnvironmentVariables(app *v1alpha1.Application) []corev1.EnvVar {
	envVars := []corev1.EnvVar{}

	// Add user-defined environment variables
	for key, value := range app.Spec.Env {
		envVars = append(envVars, corev1.EnvVar{Name: key, Value: value})
	}

	// Add infrastructure connection details (environment-aware)
	if app.Status.DatabaseEndpoint != "" {
		dbName := "webapp"
		if app.Spec.Infrastructure.PostgreSQL != nil && app.Spec.Infrastructure.PostgreSQL.DatabaseName != "" {
			dbName = app.Spec.Infrastructure.PostgreSQL.DatabaseName
		}
		
		if app.Status.DatabaseEnvironment == v1alpha1.EnvironmentLocal {
			envVars = append(envVars, corev1.EnvVar{
				Name:  "DATABASE_URL",
				Value: fmt.Sprintf("postgres://appuser:localpassword@%s/%s", app.Status.DatabaseEndpoint, dbName),
			})
		} else {
			envVars = append(envVars, corev1.EnvVar{
				Name:  "DATABASE_URL",
				Value: fmt.Sprintf("postgres://user:password@%s/%s", app.Status.DatabaseEndpoint, dbName),
			})
		}
	}

	if app.Status.RedisEndpoint != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "REDIS_URL",
			Value: fmt.Sprintf("redis://%s", app.Status.RedisEndpoint),
		})
	}

	if app.Status.S3BucketName != "" {
		envVars = append(envVars, corev1.EnvVar{Name: "S3_BUCKET", Value: app.Status.S3BucketName})
		
		if app.Status.S3Environment == v1alpha1.EnvironmentLocal {
			envVars = append(envVars, corev1.EnvVar{Name: "S3_ENDPOINT", Value: fmt.Sprintf("http://%s", app.Status.S3Endpoint)})
			envVars = append(envVars, corev1.EnvVar{Name: "S3_ACCESS_KEY", Value: "minioadmin"})
			envVars = append(envVars, corev1.EnvVar{Name: "S3_SECRET_KEY", Value: "minioadmin"})
		}
	}

	return envVars
}

// Keep all existing methods (createOrUpdateDeployment, createOrUpdateService, etc.)
// ... (include all the remaining methods from the previous version)

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

	if err := r.Create(ctx, deployment); err != nil {
		if errors.IsAlreadyExists(err) {
			logger.Info("📦 Deployment already exists, updating...")
			return nil
		}
		return fmt.Errorf("failed to create deployment: %w", err)
	}

	logger.Info("✅ Created Kubernetes Deployment", "replicas", app.GetReplicas())
	return nil
}

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

func (r *ApplicationController) checkApplicationReady(ctx context.Context, app *v1alpha1.Application) (bool, error) {
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, client.ObjectKey{Name: app.Name, Namespace: app.Namespace}, deployment)
	if err != nil {
		return false, err
	}

	if deployment.Status.ReadyReplicas == app.GetReplicas() {
		app.Status.ReadyReplicas = deployment.Status.ReadyReplicas
		return true, nil
	}

	app.Status.ReadyReplicas = deployment.Status.ReadyReplicas
	return false, nil
}

func (r *ApplicationController) updateApplicationStatus(ctx context.Context, app *v1alpha1.Application) (ctrl.Result, error) {
	if err := r.Status().Update(ctx, app); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update Application status: %w", err)
	}
	return ctrl.Result{}, nil
}

func (r *ApplicationController) updateApplicationStatusOnly(ctx context.Context, app *v1alpha1.Application) error {
	if err := r.Status().Update(ctx, app); err != nil {
		return fmt.Errorf("failed to update Application status: %w", err)
	}
	return nil
}

func (r *ApplicationController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Application{}).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Complete(r)
}