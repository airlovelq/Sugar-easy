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

package controllers

import (
	"context"
	"strconv"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sugarv1 "DistributeTrainJob/api/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type UpdateIgnorePredicate struct {
	predicate.Funcs
}

func (rl *UpdateIgnorePredicate) Update(e event.UpdateEvent) bool {
	_, ok1 := e.ObjectOld.(*sugarv1.DistributeTrainJob)
	_, ok2 := e.ObjectNew.(*sugarv1.DistributeTrainJob)
	if ok1 && ok2 {
		return false
	} else {
		return true
	}
}

// DistributeTrainJobReconciler reconciles a DistributeTrainJob object
type DistributeTrainJobReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=sugar.sugar.com,resources=distributetrainjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sugar.sugar.com,resources=distributetrainjobs/status,verbs=get;update;patch

func (r *DistributeTrainJobReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithName(req.NamespacedName.String())
	// (WithValues("distributetrainjob", req.NamespacedName)

	// your logic here
	// Get This Job
	disTrainJob := &sugarv1.DistributeTrainJob{}
	err := r.Get(ctx, req.NamespacedName, disTrainJob)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
		// if err.(*errors.StatusError).Status().Code == 404 {
		// 	return ctrl.Result{}, nil
		// } else {
		// 	return ctrl.Result{Requeue: true}, err
		// }
	}
	r.Log.Info("Get Distribute Train Job Success")
	// defer r.Update(ctx, disTrainJob)
	// defer r.Status().Update(ctx, disTrainJob)

	if err = r.finalize(ctx, disTrainJob); err != nil {
		return ctrl.Result{}, err
	}
	if r.getService(ctx, req.NamespacedName) == nil {
		r.Log.Info("Service Nod Found, Create Service")
		err = r.createService(ctx, disTrainJob)
		// err = r.Create(ctx, service)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	relatedPods := &corev1.PodList{}
	err = r.getPods(ctx, disTrainJob, relatedPods)
	if err != nil {
		r.Log.Info("Get Pods Failed")
		return ctrl.Result{Requeue: true}, err
	}
	r.Log.Info("Get Pods Success")

	disTrainJob.Status.ErrorReplicas = 0
	disTrainJob.Status.RunningReplicas = 0
	disTrainJob.Status.CompleteReplicas = 0
	disTrainJob.Status.PendingReplicas = 0

	// var errorReplicas int32 = 0
	// var runningReplicas int32 = 0
	// var completeReplicas int32 = 0
	// var pendingReplicas int32 = 0
	// Calculate number of pod in each status
	for _, pod := range relatedPods.Items {
		if pod.Status.Phase == corev1.PodFailed {
			disTrainJob.Status.ErrorReplicas++
		} else if pod.Status.Phase == corev1.PodRunning {
			disTrainJob.Status.RunningReplicas++
		} else if pod.Status.Phase == corev1.PodSucceeded {
			disTrainJob.Status.CompleteReplicas++
		} else if pod.Status.Phase == corev1.PodPending {
			disTrainJob.Status.PendingReplicas++
		}
	}

	// disTrainJob.Status.ErrorReplicas = errorReplicas
	// disTrainJob.Status.RunningReplicas = runningReplicas
	// disTrainJob.Status.CompleteReplicas = completeReplicas
	// disTrainJob.Status.PendingReplicas = pendingReplicas
	if disTrainJob.Status.Status == sugarv1.TERMINATED {
		return ctrl.Result{}, nil
	}
	if disTrainJob.Status.Status == sugarv1.SUCCEED || disTrainJob.Status.Status == sugarv1.FAILED {
		err = r.deletePods(ctx, relatedPods)
		if err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		err = r.Status().Update(ctx, disTrainJob)
		return ctrl.Result{}, nil
		// disTrainJob.ObjectMeta.DeletionTimestamp = nil
		// err = r.Update(ctx, disTrainJob)
		// err = r.Status().Update(ctx, disTrainJob)
		// return ctrl.Result{}, err
	}

	// logic in different status
	if disTrainJob.Status.ErrorReplicas > 0 {
		disTrainJob.Status.Status = sugarv1.FAILED
		r.Log.Info("Some Pod Failed, Distribute Train Job Failed")
		r.Recorder.Eventf(disTrainJob, corev1.EventTypeWarning, "Error", "Some Pod Failed, Distribute Train Job Failed")
		// r.deletePods(ctx, relatedPods)
		r.Status().Update(ctx, disTrainJob)
		// r.Delete(ctx, disTrainJob)
		// r.Update(ctx, disTrainJob)
		return ctrl.Result{}, nil
	}
	if disTrainJob.Status.CompleteReplicas == disTrainJob.Spec.Replicas {
		disTrainJob.Status.Status = sugarv1.SUCCEED
		r.Log.Info("All Pod Complete, Distribute Train Job Success")
		r.Recorder.Eventf(disTrainJob, corev1.EventTypeNormal, "Info", "All Pod Complete, Distribute Train Job Success")
		// disTrainJob.ObjectMeta.DeletionTimestamp = nil
		// r.Update(ctx, disTrainJob)
		r.Status().Update(ctx, disTrainJob)
		// r.Delete(ctx, disTrainJob)
		return ctrl.Result{}, nil
	}

	readyCount := int32(len(relatedPods.Items))
	// r.Log.Info("job success" + strconv.Itoa(int(readyCount)))
	if readyCount < disTrainJob.Spec.Replicas {
		if disTrainJob.Status.Status == sugarv1.RUNNING {
			// 一般是手动删除了pod引起
			disTrainJob.Status.Status = sugarv1.FAILED
			// r.Status().Update(ctx, disTrainJob)
		} else {
			disTrainJob.Status.Status = sugarv1.INIT
			err := r.createPod(ctx, disTrainJob, int(readyCount))
			// err := r.Create(ctx, newPod, &client.CreateOptions{})
			if err != nil {
				return ctrl.Result{}, err
			}
			// panic(err)
			// } else {
			// 	r.Log.Info("Create Pod Success", "Index", readyCount)
			// 	if err = ctrl.SetControllerReference(disTrainJob, newPod, r.Scheme); err != nil {
			// 		return ctrl.Result{}, err
			// 	}
			// }
			// r.Status().Update(ctx, disTrainJob)
		}
		// r.Status().Update(ctx, disTrainJob)
		// for ; readyCount < disTrainJob.Spec.Replicas; readyCount++ {
		// 	r.Log.Info("create success")
		// 	newPod := r.newPod(disTrainJob)
		// 	err := r.Create(context.TODO(), newPod, &client.CreateOptions{})
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// 	// r.Update(ctx, newPod)
		// 	// disTrainJob.Status.RunningReplicas++
		// }
	} else {
		if disTrainJob.Status.Status != sugarv1.RUNNING {
			disTrainJob.Status.Status = sugarv1.RUNNING
			r.Log.Info("All Pod Running, Distribute Job Running")
			r.Recorder.Eventf(disTrainJob, corev1.EventTypeNormal, "Info", "All Pod Running, Distribute Job Running")
		}
		// r.Status().Update(ctx, disTrainJob)
	}
	r.Status().Update(ctx, disTrainJob)
	return ctrl.Result{}, nil
}

func (r *DistributeTrainJobReconciler) createService(ctx context.Context, job *sugarv1.DistributeTrainJob) error {
	// bc := true
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      job.Name,
			Namespace: job.Namespace,
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
		Spec: corev1.ServiceSpec{
			ClusterIP: corev1.ClusterIPNone,
			Ports:     job.Spec.Ports,
			Selector:  job.Spec.Selector.MatchLabels,
		},
	}
	err := ctrl.SetControllerReference(job, service, r.Scheme)
	if err != nil {
		r.Log.Info("Set Service Owner Error")
		return err
	}
	err = r.Create(ctx, service)
	if err != nil {
		r.Log.Info("Create Service Error")
	}

	return err
}

func (r *DistributeTrainJobReconciler) getService(ctx context.Context, name types.NamespacedName) *corev1.Service {
	// Get pod of this job by label selector
	service := &corev1.Service{}
	err := r.Get(ctx, name, service)
	if err != nil {
		if err.(*errors.StatusError).Status().Code == 404 {
			return nil
		}
	}
	return service
}

func (r *DistributeTrainJobReconciler) deletePods(ctx context.Context, pods *corev1.PodList) error {
	var err error
	for _, pod := range pods.Items {
		err = r.Delete(ctx, &pod, &client.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *DistributeTrainJobReconciler) getPods(ctx context.Context, job *sugarv1.DistributeTrainJob, relatedPods *corev1.PodList) error {
	// Get pod of this job by label selector
	// relatedPods := &corev1.PodList{}
	// labelSelector := labels.SelectorFromValidatedSet(job.Spec.Selector.MatchLabels)
	// err := r.List(ctx, relatedPods, &client.ListOptions{LabelSelector: labelSelector})
	err := r.List(ctx, relatedPods, client.InNamespace(job.Namespace), client.MatchingFields{jobOwnerKey: job.Name})
	return err
}

func (r *DistributeTrainJobReconciler) createPod(ctx context.Context, job *sugarv1.DistributeTrainJob, no int) error {
	// bc := true
	podName := job.Name + "-" + strconv.Itoa(no)

	var portsStr string
	for i, port := range job.Spec.Ports {
		portsStr += strconv.Itoa(int(port.Port))
		if i != len(job.Spec.Ports)-1 {
			portsStr += ","
		}
	}
	var ipsStr string
	for i := 0; i < int(job.Spec.Replicas); i++ {
		ip := job.Name + "-" + strconv.Itoa(i) + "." + job.Name + "." + job.Namespace + ".svc.cluster.local"
		ipsStr += ip
		if i != int(job.Spec.Replicas-1) {
			ipsStr += ","
		}
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: job.Namespace,
			Labels:    job.Spec.Selector.MatchLabels,
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
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Hostname:      podName,
			Subdomain:     job.Name,
			Containers: []corev1.Container{
				{
					Name:            "trainjob",
					Image:           "sugar/trainer:latest",
					ImagePullPolicy: corev1.PullIfNotPresent,
					Resources:       job.Spec.Resources,
					Env: []corev1.EnvVar{
						{
							Name:  "MODEL_SAVE_PATH",
							Value: job.Spec.ModelSavePath,
						},
						{
							Name:  "MODEL_FILE",
							Value: job.Spec.ModelFile,
						},
						{
							Name:  "MODEL_CLASS",
							Value: job.Spec.ModelClass,
						},
						{
							Name:  "MODEL_CHECKPOINT",
							Value: job.Spec.ModelCheckpoint,
						},
						{
							Name:  "MODEL_PARAMS",
							Value: job.Spec.ModelParams,
						},
						{
							Name:  "TRAIN_PARAMS",
							Value: job.Spec.TrainParams,
						},
						{
							Name:  "TRAIN_DATASET",
							Value: job.Spec.TrainDataset,
						},
						{
							Name:  "VALIDATE_DATASET",
							Value: job.Spec.ValidateDataset,
						},
						{
							Name:  "MODEL_SAVE_CHECKPOINT_PATH",
							Value: job.Spec.ModelSaveCheckpointPath,
						},
						{
							Name:  "MAX_TRIALS",
							Value: strconv.Itoa(int(job.Spec.MaxTrials)),
						},
						{
							Name:  "DEST_SCORE",
							Value: job.Spec.DestScore,
						},
						{
							Name:  "USE_AUTOML",
							Value: strconv.Itoa(int(job.Spec.UseAutoml)),
						},
						{
							Name:  "DistributeTrainJobID",
							Value: strconv.Itoa(no),
						},
						{
							Name:  "DistributeTrainJobPorts",
							Value: portsStr,
						},
						{
							Name:  "DistributeTrainJobIPs",
							Value: ipsStr,
						},
					},
					VolumeMounts: []corev1.VolumeMount{

						{
							Name:      "modelfolder",
							MountPath: "/root/model",
							ReadOnly:  true,
						},
						{
							Name:      "modelparamfolder",
							MountPath: "/root/params",
							ReadOnly:  false,
						},
						{
							Name:      "trainfolder",
							MountPath: "/root/train",
							ReadOnly:  false,
						},
						{
							Name:      "datasetfolder",
							MountPath: "/root/dataset",
							ReadOnly:  false,
						},
						{
							Name:      "checkpointfolder",
							MountPath: "/root/checkpoint",
							ReadOnly:  false,
						},
					},
				},
			},

			Volumes: []corev1.Volume{
				job.Spec.ModelParamVolume,
				job.Spec.ModelVolume,
				job.Spec.TrainParamVolume,
				job.Spec.DatasetVolume,
				job.Spec.CheckpointVolume,
			},
		},

		// Spec: *job.Spec.Template.Spec.DeepCopy(),
	}
	// pod.Spec.RestartPolicy = corev1.RestartPolicyNever

	// for i := range pod.Spec.Containers {
	// 	pod.Spec.Containers[i].Env = append(pod.Spec.Containers[i].Env, corev1.EnvVar{Name: "DistributeTrainJobID", Value: strconv.Itoa(no)})
	// 	pod.Spec.Containers[i].Env = append(pod.Spec.Containers[i].Env, corev1.EnvVar{Name: "DistributeTrainJobPorts", Value: portsStr})
	// 	pod.Spec.Containers[i].Env = append(pod.Spec.Containers[i].Env, corev1.EnvVar{Name: "DistributeTrainJobIPs", Value: ipsStr})
	// }
	// pod.Spec.Hostname = podName
	// pod.Spec.Subdomain = job.Name
	err := ctrl.SetControllerReference(job, pod, r.Scheme)
	if err != nil {
		r.Log.Info("Set Pod Owner Error", "Index", no)
		return err
	}
	err = r.Create(ctx, pod)
	if err != nil {
		r.Log.Info("Create Pod Failed", "Index", no)
	}

	return err
}

func (r *DistributeTrainJobReconciler) finalize(ctx context.Context, job *sugarv1.DistributeTrainJob) error {
	// 这里我们自定义一个 finalizer 字段
	myFinalizerName := "distributetrainjob.finalizers.io"

	if job.ObjectMeta.DeletionTimestamp.IsZero() {
		//这里由于 DeletionTimestamp 是0，即没有删除操作，则不进行处理，只检查 CR 中是否含有自定义 finalizer 字符，若没有则增加。
		if !containsString(job.ObjectMeta.Finalizers, myFinalizerName) {
			// if len(job.ObjectMeta.Finalizers) == 0 {
			r.Log.Info("Register Finalizer")
			job.ObjectMeta.Finalizers = append(job.ObjectMeta.Finalizers, myFinalizerName)
			if err := r.Update(ctx, job); err != nil {
				return err
			}
		}
	} else {
		//进行预删除操作
		if containsString(job.ObjectMeta.Finalizers, myFinalizerName) {
			// if len(job.ObjectMeta.Finalizers) > 0 {
			// do something
			r.Log.Info("running finalizer")
			if job.Status.Status != sugarv1.TERMINATED && job.Status.Status != sugarv1.SUCCEED && job.Status.Status != sugarv1.FAILED {
				job.Status.Status = "TERMINATED"
				if err := r.Status().Update(ctx, job); err != nil {
					return err
				}
				r.Log.Info("Distribute Job Terminated")
			}
			// r.Delete(ctx, job)
			job.ObjectMeta.Finalizers = removeString(job.ObjectMeta.Finalizers, myFinalizerName)
			if err := r.Update(ctx, job); err != nil {
				return err
			}
		}
	}
	return nil
}

var (
	jobOwnerKey = ".metadata.controller"
	apiGVStr    = sugarv1.GroupVersion.String()
)

func (r *DistributeTrainJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(),
		&corev1.Pod{}, jobOwnerKey, func(rawObj runtime.Object) []string {
			// job owner...
			pod := rawObj.(*corev1.Pod)
			owner := metav1.GetControllerOf(pod)
			if owner == nil {
				return nil
			}
			// ... owner DistributeTrainJob...
			if owner.APIVersion != apiGVStr || owner.Kind != "DistributeTrainJob" {
				return nil
			}
			// ... DistributeTrainJob
			return []string{owner.Name}
		}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&sugarv1.DistributeTrainJob{}).
		Owns(&corev1.Pod{}).WithEventFilter(&UpdateIgnorePredicate{}).Complete(r)
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
