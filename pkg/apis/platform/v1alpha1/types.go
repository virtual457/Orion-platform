// pkg/apis/platform/v1alpha1/types.go
// Fully Kubernetes-compatible types

package v1alpha1

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Environment types
type Environment string

const (
	EnvironmentLocal Environment = "local"
	EnvironmentAWS   Environment = "aws"
	EnvironmentAuto  Environment = "auto"
)

// ApplicationSpec defines what the developer wants to deploy
type ApplicationSpec struct {
	Image    string            `json:"image"`
	Port     int32             `json:"port,omitempty"`
	Replicas int32             `json:"replicas,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
	Infrastructure InfrastructureSpec `json:"infrastructure,omitempty"`
}

// InfrastructureSpec defines external AWS resources needed
type InfrastructureSpec struct {
	Environment Environment     `json:"environment,omitempty"`
	PostgreSQL  *PostgreSQLSpec `json:"postgresql,omitempty"`
	Redis       *RedisSpec      `json:"redis,omitempty"`
	S3          *S3Spec         `json:"s3,omitempty"`
}

type PostgreSQLSpec struct {
	Environment  Environment `json:"environment,omitempty"`
	Version      string      `json:"version,omitempty"`
	InstanceType string      `json:"instanceType,omitempty"`
	Storage      int32       `json:"storage,omitempty"`
	DatabaseName string      `json:"databaseName,omitempty"`
	LocalStorage string      `json:"localStorage,omitempty"`
}

type RedisSpec struct {
	Environment Environment `json:"environment,omitempty"`
	Version     string      `json:"version,omitempty"`
	NodeType    string      `json:"nodeType,omitempty"`
	Memory      string      `json:"memory,omitempty"`
}

type S3Spec struct {
	Environment  Environment `json:"environment,omitempty"`
	BucketName   string      `json:"bucketName,omitempty"`
	Versioning   bool        `json:"versioning,omitempty"`
	LocalStorage string      `json:"localStorage,omitempty"`
}

// ApplicationStatus shows current state
type ApplicationStatus struct {
	Phase               ApplicationPhase `json:"phase,omitempty"`
	Message             string           `json:"message,omitempty"`
	ReadyReplicas       int32            `json:"readyReplicas,omitempty"`
	LastUpdated         metav1.Time      `json:"lastUpdated,omitempty"`
	InfrastructureReady bool             `json:"infrastructureReady,omitempty"`
	DatabaseEndpoint    string           `json:"databaseEndpoint,omitempty"`
	DatabaseEnvironment Environment      `json:"databaseEnvironment,omitempty"`
	RedisEndpoint       string           `json:"redisEndpoint,omitempty"`
	RedisEnvironment    Environment      `json:"redisEnvironment,omitempty"`
	S3BucketName        string           `json:"s3BucketName,omitempty"`
	S3Endpoint          string           `json:"s3Endpoint,omitempty"`
	S3Environment       Environment      `json:"s3Environment,omitempty"`
}

type ApplicationPhase string

const (
	PhasePending           ApplicationPhase = "Pending"
	PhaseProvisioningInfra ApplicationPhase = "ProvisioningInfrastructure"
	PhaseDeploying         ApplicationPhase = "Deploying"
	PhaseReady             ApplicationPhase = "Ready"
	PhaseFailed            ApplicationPhase = "Failed"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=app
// Application is our main Custom Resource
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

// GetObjectKind implements runtime.Object interface
func (app *Application) GetObjectKind() schema.ObjectKind {
	return &app.TypeMeta
}

// GetObjectKind implements runtime.Object interface  
func (appList *ApplicationList) GetObjectKind() schema.ObjectKind {
	return &appList.TypeMeta
}

// DeepCopyObject implements runtime.Object interface
func (app *Application) DeepCopyObject() runtime.Object {
	if c := app.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyObject implements runtime.Object interface
func (appList *ApplicationList) DeepCopyObject() runtime.Object {
	if c := appList.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopy creates a deep copy of Application
func (app *Application) DeepCopy() *Application {
	if app == nil {
		return nil
	}
	out := new(Application)
	app.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (app *Application) DeepCopyInto(out *Application) {
	*out = *app
	out.TypeMeta = app.TypeMeta
	app.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	app.Spec.DeepCopyInto(&out.Spec)
	app.Status.DeepCopyInto(&out.Status)
}

// DeepCopy creates a deep copy of ApplicationList
func (appList *ApplicationList) DeepCopy() *ApplicationList {
	if appList == nil {
		return nil
	}
	out := new(ApplicationList)
	appList.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies all properties into another ApplicationList
func (appList *ApplicationList) DeepCopyInto(out *ApplicationList) {
	*out = *appList
	out.TypeMeta = appList.TypeMeta
	appList.ListMeta.DeepCopyInto(&out.ListMeta)
	if appList.Items != nil {
		in, out := &appList.Items, &out.Items
		*out = make([]Application, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopyInto for ApplicationSpec
func (spec *ApplicationSpec) DeepCopyInto(out *ApplicationSpec) {
	*out = *spec
	if spec.Env != nil {
		in, out := &spec.Env, &out.Env
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	spec.Infrastructure.DeepCopyInto(&out.Infrastructure)
}

// DeepCopyInto for InfrastructureSpec
func (infra *InfrastructureSpec) DeepCopyInto(out *InfrastructureSpec) {
	*out = *infra
	if infra.PostgreSQL != nil {
		in, out := &infra.PostgreSQL, &out.PostgreSQL
		*out = new(PostgreSQLSpec)
		**out = **in
	}
	if infra.Redis != nil {
		in, out := &infra.Redis, &out.Redis
		*out = new(RedisSpec)
		**out = **in
	}
	if infra.S3 != nil {
		in, out := &infra.S3, &out.S3
		*out = new(S3Spec)
		**out = **in
	}
}

// DeepCopyInto for ApplicationStatus
func (status *ApplicationStatus) DeepCopyInto(out *ApplicationStatus) {
	*out = *status
}

// Business logic methods with Kubernetes-compatible time handling
func (app *Application) UpdateStatus(phase ApplicationPhase, message string) {
	app.Status.Phase = phase
	app.Status.Message = message
	app.Status.LastUpdated = metav1.NewTime(time.Now())
}

func (app *Application) IsReady() bool {
	return app.Status.Phase == PhaseReady && app.Status.ReadyReplicas > 0
}

func (app *Application) NeedsDatabase() bool {
	return app.Spec.Infrastructure.PostgreSQL != nil
}

func (app *Application) NeedsCache() bool {
	return app.Spec.Infrastructure.Redis != nil
}

func (app *Application) NeedsStorage() bool {
	return app.Spec.Infrastructure.S3 != nil
}

func (app *Application) GetDatabaseEnvironment() Environment {
	if app.Spec.Infrastructure.PostgreSQL != nil && app.Spec.Infrastructure.PostgreSQL.Environment != "" {
		return app.Spec.Infrastructure.PostgreSQL.Environment
	}
	if app.Spec.Infrastructure.Environment != "" {
		return app.Spec.Infrastructure.Environment
	}
	return EnvironmentAuto
}

func (app *Application) GetRedisEnvironment() Environment {
	if app.Spec.Infrastructure.Redis != nil && app.Spec.Infrastructure.Redis.Environment != "" {
		return app.Spec.Infrastructure.Redis.Environment
	}
	if app.Spec.Infrastructure.Environment != "" {
		return app.Spec.Infrastructure.Environment
	}
	return EnvironmentAuto
}

func (app *Application) GetS3Environment() Environment {
	if app.Spec.Infrastructure.S3 != nil && app.Spec.Infrastructure.S3.Environment != "" {
		return app.Spec.Infrastructure.S3.Environment
	}
	if app.Spec.Infrastructure.Environment != "" {
		return app.Spec.Infrastructure.Environment
	}
	return EnvironmentAuto
}

func (app *Application) IsLocalDatabase() bool {
	env := app.GetDatabaseEnvironment()
	return env == EnvironmentLocal || (env == EnvironmentAuto && app.isLocalEnvironment())
}

func (app *Application) IsLocalRedis() bool {
	env := app.GetRedisEnvironment()
	return env == EnvironmentLocal || (env == EnvironmentAuto && app.isLocalEnvironment())
}

func (app *Application) IsLocalS3() bool {
	env := app.GetS3Environment()
	return env == EnvironmentLocal || (env == EnvironmentAuto && app.isLocalEnvironment())
}

func (app *Application) isLocalEnvironment() bool {
	return true // For now, default to local
}

func (app *Application) ValidateSpec() error {
	if app.Spec.Image == "" {
		return fmt.Errorf("image is required")
	}
	if app.Spec.Port != 0 && (app.Spec.Port < 1 || app.Spec.Port > 65535) {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if app.Spec.Replicas < 0 {
		return fmt.Errorf("replicas cannot be negative")
	}
	return nil
}

func (app *Application) GetReplicas() int32 {
	if app.Spec.Replicas <= 0 {
		return 1
	}
	return app.Spec.Replicas
}

func (app *Application) GetPort() int32 {
	if app.Spec.Port <= 0 {
		return 8080
	}
	return app.Spec.Port
}

func (app *Application) GetInfrastructureSummary() string {
	var components []string
	
	if app.NeedsDatabase() {
		env := app.GetDatabaseEnvironment()
		if app.IsLocalDatabase() {
			components = append(components, fmt.Sprintf("PostgreSQL (local:%s)", env))
		} else {
			components = append(components, fmt.Sprintf("PostgreSQL (AWS:%s)", env))
		}
	}
	
	if app.NeedsCache() {
		env := app.GetRedisEnvironment()
		if app.IsLocalRedis() {
			components = append(components, fmt.Sprintf("Redis (local:%s)", env))
		} else {
			components = append(components, fmt.Sprintf("Redis (AWS:%s)", env))
		}
	}
	
	if app.NeedsStorage() {
		env := app.GetS3Environment()
		if app.IsLocalS3() {
			components = append(components, fmt.Sprintf("S3/MinIO (local:%s)", env))
		} else {
			components = append(components, fmt.Sprintf("S3 (AWS:%s)", env))
		}
	}
	
	if len(components) == 0 {
		return "No external infrastructure"
	}
	return fmt.Sprintf("Infrastructure: %v", components)
}