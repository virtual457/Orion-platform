package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/virtual457/orion-platform/pkg/apis/platform/v1alpha1"
	"github.com/virtual457/orion-platform/pkg/controllers"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	printBanner()

	// Check if we're running in development mode (no kubeconfig)
	if isDevelopmentMode() {
		setupLog.Info("🚧 Running in development mode - simulating controller")
		runDevelopmentMode()
		return
	}

	// Production mode - real Kubernetes controller
	runProductionMode(metricsAddr, probeAddr, enableLeaderElection)
}

func printBanner() {
	fmt.Println("🚀 =====================================================")
	fmt.Println("🚀 ORION PLATFORM - KUBERNETES OPERATOR")
	fmt.Println("🚀 =====================================================")
	fmt.Printf("🚀 Version: 0.2.0\n")
	fmt.Printf("🚀 Build Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("🚀 Environment: Production Kubernetes Controller\n")
	fmt.Println("🚀 =====================================================")
}

// isDevelopmentMode checks if we're running without Kubernetes
func isDevelopmentMode() bool {
	// If kubeconfig is not available, run in dev mode
	_, err := ctrl.GetConfig()
	return err != nil
}

// runDevelopmentMode simulates the controller for local testing
func runDevelopmentMode() {
	setupLog.Info("🧪 DEVELOPMENT MODE - Simulating Kubernetes Controller")

	// Create a sample application to show what the controller would do
	app := &v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample-web-app",
			Namespace: "default",
		},
		Spec: v1alpha1.ApplicationSpec{
			Image:    "nginx:latest",
			Port:     80,
			Replicas: 3,
			Env: map[string]string{
				"ENV":       "development",
				"LOG_LEVEL": "debug",
			},
			Infrastructure: v1alpha1.InfrastructureSpec{
				PostgreSQL: &v1alpha1.PostgreSQLSpec{
					Version:      "14.9",
					InstanceType: "db.t3.micro",
					Storage:      20,
					DatabaseName: "webapp",
				},
				Redis: &v1alpha1.RedisSpec{
					Version:  "7.0",
					NodeType: "cache.t3.micro",
				},
			},
		},
	}

	// Simulate the controller reconciliation loop
	simulateReconciliation(app)
}

// simulateReconciliation shows what the real controller would do
func simulateReconciliation(app *v1alpha1.Application) {
	setupLog.Info("🔄 Starting reconciliation simulation", "app", app.Name)

	// Phase 1: Validation
	setupLog.Info("📋 Validating application specification")
	if err := app.ValidateSpec(); err != nil {
		setupLog.Error(err, "❌ Validation failed")
		return
	}
	setupLog.Info("✅ Application specification valid")

	// Phase 2: Infrastructure Provisioning
	setupLog.Info("🏗️ Simulating AWS infrastructure provisioning")
	app.UpdateStatus(v1alpha1.PhaseProvisioningInfra, "Provisioning PostgreSQL and Redis")

	// Simulate infrastructure work
	time.Sleep(2 * time.Second)

	app.Status.InfrastructureReady = true
	app.Status.DatabaseEndpoint = "webapp-db.cluster-xyz.us-west-2.rds.amazonaws.com"
	app.Status.RedisEndpoint = "webapp-cache.xyz.cache.amazonaws.com"

	setupLog.Info("✅ Infrastructure provisioning complete",
		"database", app.Status.DatabaseEndpoint,
		"cache", app.Status.RedisEndpoint)

	// Phase 3: Kubernetes Deployment
	setupLog.Info("🚀 Simulating Kubernetes deployment creation")
	app.UpdateStatus(v1alpha1.PhaseDeploying, "Creating Deployment and Service")

	// Simulate deployment work
	time.Sleep(2 * time.Second)

	setupLog.Info("📦 Created Kubernetes Deployment", "replicas", app.GetReplicas())
	setupLog.Info("🌐 Created Kubernetes Service", "port", app.GetPort())

	// Phase 4: Ready
	app.Status.ReadyReplicas = app.GetReplicas()
	app.UpdateStatus(v1alpha1.PhaseReady, "All replicas ready and serving traffic")

	setupLog.Info("🎉 Application deployment complete!",
		"phase", app.Status.Phase,
		"readyReplicas", app.Status.ReadyReplicas,
		"isReady", app.IsReady())

	// Show final status
	fmt.Println("\n📊 FINAL APPLICATION STATUS:")
	fmt.Printf("   Name: %s\n", app.Name)
	fmt.Printf("   Phase: %s\n", app.Status.Phase)
	fmt.Printf("   Message: %s\n", app.Status.Message)
	fmt.Printf("   Ready: %t\n", app.IsReady())
	fmt.Printf("   Replicas: %d/%d\n", app.Status.ReadyReplicas, app.GetReplicas())
	fmt.Printf("   Database: %s\n", app.Status.DatabaseEndpoint)
	fmt.Printf("   Cache: %s\n", app.Status.RedisEndpoint)

	fmt.Println("\n🎯 WHAT HAPPENS IN PRODUCTION:")
	fmt.Println("   • Controller watches for Application resources")
	fmt.Println("   • Provisions real AWS RDS and ElastiCache")
	fmt.Println("   • Creates actual Kubernetes Deployments")
	fmt.Println("   • Manages full application lifecycle")
	fmt.Println("   • Handles failures and scaling automatically")

	fmt.Println("\n🚀 Next: Deploy to real Kubernetes cluster!")
}

// runProductionMode runs the real Kubernetes controller
func runProductionMode(metricsAddr, probeAddr string, enableLeaderElection bool) {
	setupLog.Info("🎯 PRODUCTION MODE - Starting Kubernetes Controller Manager")

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "orion-platform-controller",
	})
	if err != nil {
		setupLog.Error(err, "❌ Unable to start manager")
		os.Exit(1)
	}

	// Setup the Application controller
	if err = (&controllers.ApplicationController{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "❌ Unable to create controller", "controller", "Application")
		os.Exit(1)
	}

	// Setup health checks
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "❌ Unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "❌ Unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("🚀 Starting controller manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "❌ Problem running manager")
		os.Exit(1)
	}
}