// cmd/operator/main.go
// Fixed scheme registration and initialization

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

	platformv1alpha1 "github.com/virtual457/orion-platform/pkg/apis/platform/v1alpha1"
	"github.com/virtual457/orion-platform/pkg/controllers"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	// Add standard Kubernetes types to scheme
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	
	// Add our custom types to scheme
	utilruntime.Must(platformv1alpha1.AddToScheme(scheme))
	
	setupLog.Info("Scheme initialized", "groups", scheme.AllKnownTypes())
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
		setupLog.Info("Running in development mode - simulating controller")
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
	fmt.Printf("🚀 Version: 0.3.0\n")
	fmt.Printf("🚀 Build Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("🚀 Environment: Production Kubernetes Controller\n")
	fmt.Println("🚀 =====================================================")
}

// isDevelopmentMode checks if we're running without Kubernetes
func isDevelopmentMode() bool {
	_, err := ctrl.GetConfig()
	return err != nil
}

// runDevelopmentMode simulates the controller for local testing
func runDevelopmentMode() {
	setupLog.Info("DEVELOPMENT MODE - Demonstrating Smart Environment Selection")

	// Create sample application for demonstration
	app := &platformv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample-web-app",
			Namespace: "default",
		},
		Spec: platformv1alpha1.ApplicationSpec{
			Image:    "nginx:latest",
			Port:     80,
			Replicas: 3,
			Env: map[string]string{
				"ENV":       "development",
				"LOG_LEVEL": "debug",
			},
			Infrastructure: platformv1alpha1.InfrastructureSpec{
				Environment: platformv1alpha1.EnvironmentLocal,
				PostgreSQL: &platformv1alpha1.PostgreSQLSpec{
					Version:      "14.9",
					InstanceType: "db.t3.micro",
					Storage:      20,
					DatabaseName: "webapp",
				},
				Redis: &platformv1alpha1.RedisSpec{
					Version:  "7.0",
					NodeType: "cache.t3.micro",
				},
			},
		},
	}

	// Simulate reconciliation
	simulateReconciliation(app)
}

// simulateReconciliation shows what the real controller would do
func simulateReconciliation(app *platformv1alpha1.Application) {
	setupLog.Info("Starting reconciliation simulation", "app", app.Name)

	// Validation
	setupLog.Info("Validating application specification")
	if err := app.ValidateSpec(); err != nil {
		setupLog.Error(err, "Validation failed")
		return
	}
	setupLog.Info("Application specification valid")

	// Infrastructure provisioning
	setupLog.Info("Smart infrastructure provisioning")
	app.UpdateStatus(platformv1alpha1.PhaseProvisioningInfra, "Provisioning infrastructure")

	time.Sleep(2 * time.Second)

	// Simulate infrastructure ready
	app.Status.InfrastructureReady = true
	if app.IsLocalDatabase() {
		app.Status.DatabaseEndpoint = fmt.Sprintf("%s-postgres:5432", app.Name)
		app.Status.DatabaseEnvironment = platformv1alpha1.EnvironmentLocal
	}
	if app.IsLocalRedis() {
		app.Status.RedisEndpoint = fmt.Sprintf("%s-redis:6379", app.Name)
		app.Status.RedisEnvironment = platformv1alpha1.EnvironmentLocal
	}

	setupLog.Info("Infrastructure provisioning complete",
		"database", app.Status.DatabaseEndpoint,
		"cache", app.Status.RedisEndpoint)

	// Application deployment
	setupLog.Info("Creating application deployment")
	app.UpdateStatus(platformv1alpha1.PhaseDeploying, "Creating Kubernetes resources")

	time.Sleep(2 * time.Second)

	setupLog.Info("Kubernetes resources created", "replicas", app.GetReplicas())

	// Ready
	app.Status.ReadyReplicas = app.GetReplicas()
	app.UpdateStatus(platformv1alpha1.PhaseReady, "All replicas ready")

	setupLog.Info("Application deployment complete",
		"phase", app.Status.Phase,
		"ready", app.IsReady())

	// Final status
	fmt.Println("\n📊 FINAL STATUS:")
	fmt.Printf("   Name: %s\n", app.Name)
	fmt.Printf("   Phase: %s\n", app.Status.Phase)
	fmt.Printf("   Ready: %t\n", app.IsReady())
	fmt.Printf("   Database: %s (%s)\n", app.Status.DatabaseEndpoint, app.Status.DatabaseEnvironment)
	fmt.Printf("   Cache: %s (%s)\n", app.Status.RedisEndpoint, app.Status.RedisEnvironment)

	fmt.Println("\n🚀 Ready to work with real Kubernetes cluster!")
}

// runProductionMode runs the real Kubernetes controller
func runProductionMode(metricsAddr, probeAddr string, enableLeaderElection bool) {
	setupLog.Info("PRODUCTION MODE - Starting Kubernetes Controller Manager")

	// Create manager with proper scheme
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:           scheme,
		LeaderElection:   enableLeaderElection,
		LeaderElectionID: "orion-platform-controller",
	})
	if err != nil {
		setupLog.Error(err, "Unable to start manager")
		os.Exit(1)
	}

	// Setup the Application controller with proper client
	if err = (&controllers.ApplicationController{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "Application")
		os.Exit(1)
	}

	// Setup health checks
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "Unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "Unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("Starting controller manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "Problem running manager")
		os.Exit(1)
	}
}