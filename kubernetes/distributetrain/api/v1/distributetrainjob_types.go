/*


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
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DistributeTrainJobSpec defines the desired state of DistributeTrainJob
type DistributeTrainJobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Number of desired pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Label selector for pods. Existing ReplicaSets whose pods are
	// selected by this will be the ones affected by this deployment.
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// Volume For Model Folder
	ModelVolume apiv1.Volume `json:"modelvolume"`

	// Volume For Model Param
	ModelParamVolume apiv1.Volume `json:"modelparamvolume"`

	// Volume For Train
	// TrainParamVolume apiv1.Volume `json:"trainvolume"`

	// Volume For Checkpoint
	LogVolume apiv1.Volume `json:"logvolume"`

	// Volume For Dataset
	DatasetVolume apiv1.Volume `json:"datasetvolume"`

	// Resources
	Resources apiv1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`

	// Template describes the pods that will be created.
	// Template apiv1.PodTemplateSpec `json:"template,omitempty"`

	// Ports for Train
	Ports []apiv1.ServicePort `json:"ports,omitempty"`

	//
	ModelSavePath string `json:"modelsavepath"`

	//ModelFolder
	ModelFolder string `json:"modelfolder"`

	//ModelFile
	ModelFile string `json:"modelfile"`

	//
	ModelClass string `json:"modelclass"`

	//
	ModelCheckpoint string `json:"modelcheckpoint"`

	//ModelParams
	ModelParams string `json:"modelparams"`

	//TRAIN_PARAMS
	TrainParams string `json:"trainparams"`

	//
	TrainDataset string `json:"traindataset"`

	//
	ValidateDataset string `json:"validatedataset"`

	//
	// ModelSaveCheckpointPath string `json:"modelsavecheckpointpath"`

	//
	UseAutoml int32 `json:"useautoml"`

	//
	DestScore string `json:"destscore"`

	//
	MaxTrials int32 `json:"maxtrials"`

	//
	StatusInformHook *string `json:"statusinformhook"`
}

// DistributeTrainJobStatusState 声明
type DistributeTrainJobStatusState string

const (
	TERMINATED DistributeTrainJobStatusState = "TERMINATED"
	FAILED     DistributeTrainJobStatusState = "FAILED"
	SUCCEED    DistributeTrainJobStatusState = "SUCCEED"
	INIT       DistributeTrainJobStatusState = "INIT"
	RUNNING    DistributeTrainJobStatusState = "RUNNING"
)

// DistributeTrainJobStatus defines the observed state of DistributeTrainJob
type DistributeTrainJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Replicas        int32 `json:"replica,omitempty"`
	PendingReplicas int32 `json:"pendingreplicas"`

	RunningReplicas int32 `json:"runningreplicas"`

	ErrorReplicas int32 `json:"errorreplicas"`

	CompleteReplicas int32 `json:"completereplicas"`

	Status DistributeTrainJobStatusState `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.pendingreplicas",name=PendingReplicas,type=integer
// +kubebuilder:printcolumn:JSONPath=".status.runningreplicas",name=RunningReplicas,type=integer
// +kubebuilder:printcolumn:JSONPath=".status.errorreplicas",name=ErrorReplicas,type=integer
// +kubebuilder:printcolumn:JSONPath=".status.completereplicas",name=CompleteReplicas,type=integer
// +kubebuilder:printcolumn:JSONPath=".status.status",name=Status,type=string

// DistributeTrainJob is the Schema for the distributetrainjobs API
type DistributeTrainJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DistributeTrainJobSpec   `json:"spec,omitempty"`
	Status DistributeTrainJobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DistributeTrainJobList contains a list of DistributeTrainJob
type DistributeTrainJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DistributeTrainJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DistributeTrainJob{}, &DistributeTrainJobList{})
}
