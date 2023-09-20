package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	stackv1alpha1 "github.com/zncdata-labs/redis-operator/api/v1alpha1"
)

// make pvc
func makePVC(ctx context.Context, instance *stackv1alpha1.Redis, schema *runtime.Scheme) *corev1.PersistentVolumeClaim {
	logger := log.FromContext(ctx)
	labels := instance.GetLabels()
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetPvcName(),
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: instance.Spec.Persistence.StorageClass,
			AccessModes:      instance.Spec.Persistence.AccessModes,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: instance.Spec.Persistence.GetSize(),
				},
			},
			VolumeMode: instance.Spec.Persistence.VolumeMode,
		},
	}
	err := ctrl.SetControllerReference(instance, pvc, schema)
	if err != nil {
		logger.Error(err, "Failed to set controller reference")
		return nil
	}
	return pvc
}

// reconcilePVC
func (r *RedisReconciler) reconcilePVC(ctx context.Context, instance *stackv1alpha1.Redis) error {
	logger := log.FromContext(ctx)
	pvc := &corev1.PersistentVolumeClaim{}
	err := r.Client.Get(ctx, types.NamespacedName{Namespace: instance.Namespace, Name: instance.GetPvcName()}, pvc)
	if err != nil && errors.IsNotFound(err) {
		pvc := makePVC(ctx, instance, r.Scheme)
		logger.Info("Creating a new PVC", "PVC.Namespace", pvc.Namespace, "PVC.Name", pvc.Name)
		err := r.Client.Create(ctx, pvc)
		if err != nil {
			return err
		}
	} else if err != nil {
		logger.Error(err, "Failed to get PVC")
		return err
	}
	return nil
}

// make service
func makeService(ctx context.Context, instance *stackv1alpha1.Redis, schema *runtime.Scheme) *corev1.Service {
	labels := instance.GetLabels()
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.Name,
			Namespace:   instance.Namespace,
			Labels:      labels,
			Annotations: instance.Spec.Service.Annotations,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:     instance.Spec.Service.Port,
					Name:     "http",
					Protocol: "TCP",
				},
			},
			Selector: labels,
			Type:     instance.Spec.Service.Type,
		},
	}
	err := ctrl.SetControllerReference(instance, svc, schema)
	if err != nil {
		return nil
	}
	return svc
}

func (r *RedisReconciler) reconcileService(ctx context.Context, instance *stackv1alpha1.Redis) error {
	logger := log.FromContext(ctx)
	obj := makeService(ctx, instance, r.Scheme)
	if obj == nil {
		return nil
	}

	if err := CreateOrUpdate(ctx, r.Client, obj); err != nil {
		logger.Error(err, "Failed to create or update service")
		return err
	}
	return nil
}

func makeDeployment(ctx context.Context, instance *stackv1alpha1.Redis, schema *runtime.Scheme) *appsv1.Deployment {
	labels := instance.GetLabels()
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &instance.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            instance.Name,
							Image:           instance.GetImageTag(),
							ImagePullPolicy: instance.GetImagePullPolicy(),
							Args: []string{
								"/opt/bitnami/scripts/start-scripts/start-node.sh",
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 18080,
									Name:          "http",
									Protocol:      "TCP",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      instance.GetNameWithSuffix("-data"),
									MountPath: "/data",
								},
							},
						},
					},
					Tolerations: instance.Spec.Tolerations,
					Volumes: []corev1.Volume{
						{
							Name: instance.GetNameWithSuffix("-data"),
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: instance.GetPvcName(),
								},
							},
						},
					},
				},
			},
		},
	}
	err := ctrl.SetControllerReference(instance, dep, schema)
	if err != nil {
		return nil
	}
	return dep
}

func (r *RedisReconciler) reconcileDeployment(ctx context.Context, instance *stackv1alpha1.Redis) error {
	logger := log.FromContext(ctx)
	obj := makeDeployment(ctx, instance, r.Scheme)
	if obj == nil {
		return nil
	}
	if err := CreateOrUpdate(ctx, r.Client, obj); err != nil {
		logger.Error(err, "Failed to create or update deployment")
		return err
	}
	return nil
}
