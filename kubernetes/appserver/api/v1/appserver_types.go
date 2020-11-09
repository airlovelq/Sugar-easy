/*
Copyright 2020 sugar.

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

package v1

import (
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AppServerSpec defines the desired state of AppServer
type AppServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Number of min pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +optional
	ReplicasMin *int32 `json:"replicasmin,omitempty"`

	// Number of max pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 5.
	// +optional
	ReplicasMax *int32 `json:"replicasmax,omitempty"`

	// Model File Name, exclude .py
	ModelFile string `json:"modelfile"`

	// Model Param File Name
	ModelParam string `json:"modelparam"`

	// Model Class Name
	ModelClass string `json:"modelclass"`

	// Volume For Model Folder
	ModelVolume apiv1.Volume `json:"modelvolume"`

	// Volume For Model Param
	ModelParamVolume apiv1.Volume `json:"modelparamvolume"`

	// Ports for App
	Port apiv1.ServicePort `json:"port"`

	// Metrics
	Metrics []autoscalingv2.MetricSpec `json:"metrics,omitempty" protobuf:"bytes,4,rep,name=metrics"`

	// Resources
	Resources apiv1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

// DistributeTrainJobStatusState 声明
type AppServerStatusState string

const (
	TERMINATED AppServerStatusState = "TERMINATED"
	BALANCING  AppServerStatusState = "REBALANCING"
	RUNNING    AppServerStatusState = "RUNNING"
)

// AppServerStatus defines the observed state of AppServer
type AppServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ReadyReplicas int32                `json:"readyreplicas"`
	Replicas      int32                `json:"replicas"`
	Status        AppServerStatusState `json:"status"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.readyreplicas",name=ReadyReplicas,type=integer
// +kubebuilder:printcolumn:JSONPath=".status.replicas",name=Replicas,type=integer
// +kubebuilder:printcolumn:JSONPath=".status.status",name=Status,type=string

// AppServer is the Schema for the appservers API
type AppServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppServerSpec   `json:"spec,omitempty"`
	Status AppServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AppServerList contains a list of AppServer
type AppServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppServer{}, &AppServerList{})
}
