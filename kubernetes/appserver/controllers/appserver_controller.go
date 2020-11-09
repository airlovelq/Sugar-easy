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

package controllers

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/gob"
	"fmt"
	"reflect"

	"strings"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	sugarv1 "sugar.com/api/v1"
)

// AppServerReconciler reconciles a AppServer object
type AppServerReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=sugar.sugar.com,resources=appservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sugar.sugar.com,resources=appservers/status,verbs=get;update;patch

func (r *AppServerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("appserver", req.NamespacedName)
	// autoscalingv2.HorizontalPodAutoscaler
	// your logic here
	server := &sugarv1.AppServer{}
	err := r.Get(ctx, req.NamespacedName, server)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	r.Log.Info("Get App Server Success")
	var bcontinue bool = false
	if bcontinue, err = r.finalize(ctx, server); bcontinue || err != nil {
		return ctrl.Result{}, err
	}
	sha1str := strings.ToLower(getSHA1(server.Name)[0:10])
	service, err := r.getService(ctx, req.NamespacedName)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			if err = r.createService(ctx, server); err != nil {
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, err
		}
	} else {
		if err = r.updateService(ctx, server, service); err != nil {
			return ctrl.Result{}, err
		}
		// if err = r.Update(ctx, server); err != nil {
		// 	return ctrl.Result{}, err
		// }
	}

	deploy, err := r.getDeploy(ctx, server.Name+"-"+sha1str, server.Namespace)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			if err = r.createDeployment(ctx, server, server.Name+"-"+sha1str); err != nil {
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, err
		}
	} else {

		// if err = r.Update(ctx, server); err != nil {
		// 	return ctrl.Result{}, err
		// }
		// r.Get(ctx, req.NamespacedName, server)
		server.Status.Replicas = deploy.Status.Replicas
		server.Status.ReadyReplicas = deploy.Status.ReadyReplicas
		// if server.Status.Status != sugarv1.TERMINATED {
		if deploy.Status.UnavailableReplicas > 0 {
			server.Status.Status = sugarv1.BALANCING
		} else {
			server.Status.Status = sugarv1.RUNNING
		}
		if err = r.updateDeployment(ctx, server, deploy); err != nil {
			return ctrl.Result{}, err
		}
		if err = r.Status().Update(ctx, server); err != nil {
			return ctrl.Result{}, err
		}
		// }

		// err := r.Get(ctx, req.NamespacedName, server)
		// if err != nil {
		// 	return ctrl.Result{}, client.IgnoreNotFound(err)
		// }
		// deploy, _ = r.getDeploy(ctx, server.Name+"-"+sha1str, server.Namespace)

	}

	scaler, err := r.getAutoScaler(ctx, server.Name+"-"+sha1str, server.Namespace)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			if err = r.createAutoScaler(ctx, server, server.Name+"-"+sha1str); err != nil {
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, err
		}
	} else {
		if err = r.updateAutoScaler(ctx, server, scaler); err != nil {
			return ctrl.Result{}, err
		}
		// if err = r.Update(ctx, server); err != nil {
		// 	return ctrl.Result{}, err
		// }
	}
	// if err = r.Update(ctx, server); err != nil {
	// 	return ctrl.Result{}, err
	// }
	return ctrl.Result{}, nil
}

var (
	serverOwnerKey = ".metadata.controller"
	apiGVStr       = sugarv1.GroupVersion.String()
)

func (r *AppServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// if err := mgr.GetFieldIndexer().IndexField(context.Background(),
	// 	&appsv1.Deployment{}, serverOwnerKey, func(rawObj runtime.Object) []string {
	// 		// job owner...
	// 		deploy := rawObj.(*appsv1.Deployment)
	// 		owner := metav1.GetControllerOf(pod)
	// 		if owner == nil {
	// 			return nil
	// 		}
	// 		// ... owner DistributeTrainJob...
	// 		if owner.APIVersion != apiGVStr || owner.Kind != "DistributeTrainJob" {
	// 			return nil
	// 		}
	// 		// ... DistributeTrainJob
	// 		return []string{owner.Name}
	// 	}); err != nil {
	// 	return err
	// }
	return ctrl.NewControllerManagedBy(mgr).
		For(&sugarv1.AppServer{}).Owns(&appsv1.Deployment{}).Owns(&autoscalingv2.HorizontalPodAutoscaler{}).Owns(&corev1.Service{}).
		Complete(r)
}

func (r *AppServerReconciler) createDeployment(ctx context.Context, server *sugarv1.AppServer, name string) error {
	selector := make(map[string]string)
	selector["app"] = server.Name
	// requests := make(map[corev1.ResourceName]resource.Quantity)
	// cpuQuantity := resource.Quantity{}
	// cpuQuantity.Set(1)
	// requests[corev1.ResourceCPU] = cpuQuantity
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: server.Namespace,
			Labels:    selector,
			// OwnerReferences: []metav1.OwnerReference{
			// 	{
			// 		APIVersion: job.APIVersion,
			// 		Kind:       job.Kind,
			// 		Controller: &bc,
			// 		UID:        job.UID,
			// 		Name:       job.Name,
			// 	},
			// },
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: server.Spec.ReplicasMin,
			Selector: &metav1.LabelSelector{MatchLabels: selector},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: selector,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            server.Name,
							Image:           "sugar/appserver:latest",
							ImagePullPolicy: corev1.PullIfNotPresent,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "modelfolder",
									MountPath: "/root/model",
									ReadOnly:  true,
								},
								{
									Name:      "modelparamfolder",
									MountPath: "/root/params",
									ReadOnly:  true,
								},
							},
							Resources: server.Spec.Resources,
							Env: []corev1.EnvVar{
								{
									Name:  "MODEL_FILE",
									Value: server.Spec.ModelFile,
								},
								{
									Name:  "MODEL_CLASS",
									Value: server.Spec.ModelClass,
								},
								{
									Name:  "MODEL_PARAM_FILE",
									Value: server.Spec.ModelParam,
								},
								{
									Name:  "APP_PORT",
									Value: server.Spec.Port.TargetPort.String(),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						server.Spec.ModelVolume,
						server.Spec.ModelParamVolume,
					},
				},
			},
		},
	}
	err := ctrl.SetControllerReference(server, deploy, r.Scheme)
	if err != nil {
		r.Log.Error(err, "Set Deploy Owner Error")
		return err
	}
	err = r.Create(ctx, deploy)
	if err != nil {
		r.Log.Error(err, "Create Deploy Error")
		r.Recorder.Eventf(server, corev1.EventTypeWarning, "Error", "Create Deploy Failed")
	}
	r.Recorder.Eventf(server, corev1.EventTypeNormal, "Info", "Create Deploy Success")
	return err
}

func (r *AppServerReconciler) updateDeployment(ctx context.Context, server *sugarv1.AppServer, deploy *appsv1.Deployment) error {
	var bupdate bool = false
	for index, _ := range deploy.Spec.Template.Spec.Containers {
		if !reflect.DeepEqual(deploy.Spec.Template.Spec.Containers[index].Resources, server.Spec.Resources) {
			bupdate = true
			deploy.Spec.Template.Spec.Containers[index].Resources = server.Spec.Resources
		}
		for j, _ := range deploy.Spec.Template.Spec.Containers[index].Env {
			switch deploy.Spec.Template.Spec.Containers[index].Env[j].Name {
			case "MODEL_FILE":
				if deploy.Spec.Template.Spec.Containers[index].Env[j].Value != server.Spec.ModelFile {
					bupdate = true
					deploy.Spec.Template.Spec.Containers[index].Env[j].Value = server.Spec.ModelFile
				}
			case "MODEL_PARAM_FILE":
				if deploy.Spec.Template.Spec.Containers[index].Env[j].Value != server.Spec.ModelParam {
					bupdate = true
					deploy.Spec.Template.Spec.Containers[index].Env[j].Value = server.Spec.ModelParam
				}
			case "MODEL_CLASS":
				if deploy.Spec.Template.Spec.Containers[index].Env[j].Value != server.Spec.ModelClass {
					bupdate = true
					deploy.Spec.Template.Spec.Containers[index].Env[j].Value = server.Spec.ModelClass
				}
			case "APP_PORT":
				if deploy.Spec.Template.Spec.Containers[index].Env[j].Value != server.Spec.Port.TargetPort.String() {
					bupdate = true
					deploy.Spec.Template.Spec.Containers[index].Env[j].Value = server.Spec.Port.TargetPort.String()
				}
			}
		}
	}

	// err := ctrl.SetControllerReference(server, deploy, r.Scheme)
	// if err != nil {
	// 	r.Log.Info("Set Deploy Owner Error")
	// 	return err
	// }
	if !bupdate {
		return nil
	}
	err := r.Update(ctx, deploy)
	if err != nil {
		r.Log.Error(err, "Update Deploy Error")
		r.Recorder.Eventf(server, corev1.EventTypeWarning, "Error", "Update Deploy Failed")
		return err
	}
	r.Recorder.Eventf(server, corev1.EventTypeNormal, "Info", "Update Deploy Success")
	// err = r.Update(ctx, server)
	return err
}

func (r *AppServerReconciler) getDeploy(ctx context.Context, name string, namespace string) (*appsv1.Deployment, error) {
	// Get pod of this job by label selector
	deploy := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, deploy)
	return deploy, err
}

func deepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

func (r *AppServerReconciler) createService(ctx context.Context, server *sugarv1.AppServer) error {
	// bc := true.
	ports := make([]corev1.ServicePort, 0)
	// var port corev1.ServicePort
	// deepCopy(port, server.Spec.Port)
	ports = append(ports, server.Spec.Port)
	selector := make(map[string]string)
	selector["app"] = server.Name
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name,
			Namespace: server.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeNodePort,
			Ports:    ports,
			Selector: selector,
		},
	}
	err := ctrl.SetControllerReference(server, service, r.Scheme)
	if err != nil {
		r.Log.Error(err, "Set Service Owner Error")
		return err
	}
	err = r.Create(ctx, service)
	if err != nil {
		r.Log.Error(err, "Create Service Error")
		r.Recorder.Eventf(server, corev1.EventTypeWarning, "Error", "Create Service Failed")
	}
	r.Recorder.Eventf(server, corev1.EventTypeNormal, "Info", "Create Service Success")
	return err
}

func (r *AppServerReconciler) updateService(ctx context.Context, server *sugarv1.AppServer, service *corev1.Service) error {
	// bc := true
	ports := make([]corev1.ServicePort, 0)
	// var port corev1.ServicePort
	// deepCopy(port, server.Spec.Port)
	ports = append(ports, server.Spec.Port)
	// corev1.ServicePort.DeepCopy()
	// server.Spec.Port.Protocol = corev1.ProtocolTCP
	ports[0].Protocol = corev1.ProtocolTCP
	if reflect.DeepEqual(service.Spec.Ports, ports) {
		return nil
	}
	service.Spec.Ports = ports
	err := r.Update(ctx, service)
	if err != nil {
		r.Log.Error(err, "Update Service Error")
		r.Recorder.Eventf(server, corev1.EventTypeWarning, "Error", "Create Service Failed")
	}
	r.Recorder.Eventf(server, corev1.EventTypeNormal, "Info", "Create Service Success")
	// err = r.Update(ctx, server)
	return err
}

func (r *AppServerReconciler) getService(ctx context.Context, name types.NamespacedName) (*corev1.Service, error) {
	// Get pod of this job by label selector
	service := &corev1.Service{}
	err := r.Get(ctx, name, service)
	return service, err
}

func (r *AppServerReconciler) createAutoScaler(ctx context.Context, server *sugarv1.AppServer, name string) error {
	// var cpuPercentage int32 = 50
	var maxreplicas int32
	if server.Spec.ReplicasMax != nil {
		maxreplicas = *server.Spec.ReplicasMax
	} else {
		maxreplicas = 5
	}
	scaler := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: server.Namespace,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				Name:       name,
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
			MinReplicas: server.Spec.ReplicasMin,
			MaxReplicas: maxreplicas,
			Metrics:     server.Spec.Metrics,
		},
	}
	err := ctrl.SetControllerReference(server, scaler, r.Scheme)
	if err != nil {
		r.Log.Error(err, "Set Scaler Owner Error")
		return err
	}
	err = r.Create(ctx, scaler)
	if err != nil {
		r.Log.Error(err, "Create Scaler Error")
		r.Recorder.Eventf(server, corev1.EventTypeWarning, "Error", "Create Scaler Failed")
	}
	r.Recorder.Eventf(server, corev1.EventTypeNormal, "Info", "Create Scaler Success")
	return err
}

func (r *AppServerReconciler) updateAutoScaler(ctx context.Context, server *sugarv1.AppServer, scaler *autoscalingv2.HorizontalPodAutoscaler) error {
	// var cpuPercentage int32 = 50
	var maxreplicas int32
	if server.Spec.ReplicasMax != nil {
		maxreplicas = *server.Spec.ReplicasMax
	} else {
		maxreplicas = 5
	}
	// if reflect.DeepEqual(scaler.Spec.MinReplicas, server.Spec.ReplicasMin) && reflect.DeepEqual(scaler.Spec.MaxReplicas, maxreplicas) && reflect.DeepEqual(scaler.Spec.Metrics, server.Spec.Metrics) {
	// 	return nil
	// }
	scaler.Spec.MinReplicas = server.Spec.ReplicasMin
	scaler.Spec.MaxReplicas = maxreplicas
	scaler.Spec.Metrics = server.Spec.Metrics

	err := r.Update(ctx, scaler)
	if err != nil {
		r.Log.Error(err, "Update Scaler Error")
		r.Recorder.Eventf(server, corev1.EventTypeWarning, "Error", "Update Scaler Failed")
	}
	r.Recorder.Eventf(server, corev1.EventTypeNormal, "Info", "Update Scaler Success")
	// err = r.Update(ctx, server)
	return err
}

func (r *AppServerReconciler) getAutoScaler(ctx context.Context, name string, namespace string) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	// Get pod of this job by label selector
	scaler := &autoscalingv2.HorizontalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, scaler)
	return scaler, err
}

func (r *AppServerReconciler) finalize(ctx context.Context, server *sugarv1.AppServer) (bool, error) {
	// 这里我们自定义一个 finalizer 字段
	myFinalizerName := "sugarappserver.finalizers.io"

	if server.ObjectMeta.DeletionTimestamp.IsZero() {
		//这里由于 DeletionTimestamp 是0，即没有删除操作，则不进行处理，只检查 CR 中是否含有自定义 finalizer 字符，若没有则增加。
		if !containsString(server.ObjectMeta.Finalizers, myFinalizerName) {
			// if len(job.ObjectMeta.Finalizers) == 0 {
			r.Log.Info("Register Finalizer")
			server.ObjectMeta.Finalizers = append(server.ObjectMeta.Finalizers, myFinalizerName)
			if err := r.Update(ctx, server); err != nil {
				return false, err
			}
			return true, nil
		}
		return false, nil
	} else {
		//进行预删除操作
		if containsString(server.ObjectMeta.Finalizers, myFinalizerName) {
			// if len(job.ObjectMeta.Finalizers) > 0 {
			// do something
			r.Log.Info("running finalizer")
			// if server.Status.Status != sugarv1.TERMINATED {
			// 	server.Status.Status = "TERMINATED"
			// 	// if err := r.Status().Update(ctx, server); err != nil {
			// 	// 	return false, err
			// 	// }
			// 	r.Log.Info("App Server Terminated")
			// }
			// r.Delete(ctx, server)
			server.ObjectMeta.Finalizers = removeString(server.ObjectMeta.Finalizers, myFinalizerName)
			if err := r.Update(ctx, server); err != nil {
				return false, err
			}
		}
		return true, nil
	}
	return false, nil
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func getSHA1(data string) string {
	b := sha1.Sum([]byte(data))
	return fmt.Sprintf("%X", b)
}
