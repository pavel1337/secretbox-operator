/*
Copyright 2022.

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
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	secretboxv1 "github.com/pavel1337/secretbox-operator/api/v1"
)

// SecretboxReconciler reconciles a Secretbox object
type SecretboxReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=secretbox.ipvl.de,resources=secretboxes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=secretbox.ipvl.de,resources=secretboxes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=secretbox.ipvl.de,resources=secretboxes/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets/status,verbs=get
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps/status,verbs=get
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services/status,verbs=get

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *SecretboxReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the Secretbox instance
	secretbox := &secretboxv1.Secretbox{}
	err := r.Get(ctx, req.NamespacedName, secretbox)
	if errors.IsNotFound(err) {
		log.Info(fmt.Sprintf("Secretbox %s not found, must be deleted", req.NamespacedName))
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Unable to get Secretbox resource")
		return ctrl.Result{}, err
	}

	var restartDeployment bool
	// Delete existing pods if restart is required
	defer func() {
		if restartDeployment {
			r.deletePods(ctx, secretbox)
		}
	}()

	// Fetch the Secret object
	secret := &corev1.Secret{}
	err = r.Get(ctx, client.ObjectKey{Namespace: secretbox.Namespace, Name: secretbox.GetSecretName()}, secret)
	if errors.IsNotFound(err) {
		// Create the Secret object
		secret, err = r.createSecret(ctx, secretbox)
		if err != nil {
			log.Error(err, "Unable to create Secret object")
			return ctrl.Result{}, err
		}
		err = r.Create(ctx, secret)
		if err != nil {
			log.Error(err, "Unable to create Secret object")
			return ctrl.Result{}, err
		}
		restartDeployment = true
	} else if err != nil {
		log.Error(err, "Unable to get Secret object")
		return ctrl.Result{}, err
	}

	// Fetch the ConfigMap object
	configMap := &corev1.ConfigMap{}
	err = r.Get(ctx, client.ObjectKey{Namespace: secretbox.Namespace, Name: secretbox.GetConfigMapName()}, configMap)
	if errors.IsNotFound(err) {
		// Create the ConfigMap object
		configMap = r.createConfigMap(ctx, secretbox)
		err = r.Create(ctx, configMap)
		if err != nil {
			log.Error(err, "Unable to create ConfigMap object")
			return ctrl.Result{}, err
		}
		restartDeployment = true

	} else if err != nil {
		log.Error(err, "Unable to get ConfigMap object")
		return ctrl.Result{}, err
	}

	// Fetch the Deployment object
	deployment := &apps.Deployment{}
	err = r.Get(ctx, client.ObjectKey{Namespace: secretbox.Namespace, Name: secretbox.GetDeploymentName()}, deployment)
	if errors.IsNotFound(err) {
		// Create the Deployment object
		deployment = r.createDeployment(ctx, secretbox)
		err = r.Create(ctx, deployment)
		if err != nil {
			log.Error(err, "Unable to create Deployment object")

			return ctrl.Result{}, err
		}
		restartDeployment = false
	} else if err != nil {
		log.Error(err, "Unable to get Deployment object")
		return ctrl.Result{}, err
	}

	// Fetch the Service object
	service := &corev1.Service{}
	err = r.Get(ctx, client.ObjectKey{Namespace: secretbox.Namespace, Name: secretbox.GetServiceName()}, service)
	if errors.IsNotFound(err) {
		// Create the Service object
		service = r.createService(ctx, secretbox)
		err = r.Create(ctx, service)
		if err != nil {
			log.Error(err, "Unable to create Service object")
			return ctrl.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "Unable to get Service object")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// createService creates a Service object for the given Secretbox object
func (r *SecretboxReconciler) createService(ctx context.Context, secretbox *secretboxv1.Secretbox) *corev1.Service {
	log := log.FromContext(ctx)

	log.Info(fmt.Sprintf("Creating Service %s for Secretbox %s", secretbox.GetServiceName(), secretbox.Name))

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretbox.GetServiceName(),
			Namespace: secretbox.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
			},
			Selector: secretbox.GetLabels(),
		},
	}

	ctrl.SetControllerReference(secretbox, service, r.Scheme)

	return service
}

// createConfigMap creates a new ConfigMap object.
func (r *SecretboxReconciler) createConfigMap(ctx context.Context, secretbox *secretboxv1.Secretbox) *corev1.ConfigMap {
	log := log.FromContext(ctx)

	log.Info(fmt.Sprintf("Creating ConfigMap %s for Secretbox %s", secretbox.GetConfigMapName(), secretbox.Name))

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretbox.GetConfigMapName(),
			Namespace: secretbox.Namespace,
			Labels:    secretbox.GetLabels(),
		},
		Data: map[string]string{
			"LISTEN_ADDRESS":     ":8080",
			"SECRETS_STORE_TYPE": "INMEM",
			"SESSION_STORE_TYPE": "INMEM",
		},
	}

	ctrl.SetControllerReference(secretbox, configMap, r.Scheme)

	return configMap
}

// createSecret creates a new Secret object.
func (r *SecretboxReconciler) createSecret(ctx context.Context, secretbox *secretboxv1.Secretbox) (*corev1.Secret, error) {
	// Generate a random cookiesSecret.
	cookiesSecret, err := generateRandomString(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random cookiesSecret: %w", err)
	}

	// Generate random secrets encryption key.
	encryptionKey, err := generateRandomString(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random encryption key: %w", err)
	}

	log := log.FromContext(ctx)

	log.Info(fmt.Sprintf("Creating Secret %s for Secretbox %s", secretbox.GetSecretName(), secretbox.Name))

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretbox.GetSecretName(),
			Namespace: secretbox.Namespace,
		},
		Data: map[string][]byte{
			"SECRETS_ENCRYPTION_KEY": []byte(encryptionKey),
			"COOKIE_SECRET":          []byte(cookiesSecret),
		},
	}

	ctrl.SetControllerReference(secretbox, secret, r.Scheme)

	return secret, nil
}

// createDeployment creates a new Deployment object.
func (r *SecretboxReconciler) createDeployment(ctx context.Context, secretbox *secretboxv1.Secretbox) *apps.Deployment {
	log := log.FromContext(ctx)

	log.Info(fmt.Sprintf("Creating Deployment %s for Secretbox %s", secretbox.GetDeploymentName(), secretbox.Name))

	deployment := &apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretbox.GetDeploymentName(),
			Namespace: secretbox.Namespace,
			Labels:    secretbox.GetLabels(),
		},
		Spec: apps.DeploymentSpec{
			Replicas: &secretbox.Spec.Size,
			Selector: &metav1.LabelSelector{
				MatchLabels: secretbox.GetLabels(),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: secretbox.GetLabels(),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "secretbox",
							Image: "pv1337/secretbox:latest",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 8080,
								},
							},
							EnvFrom: []corev1.EnvFromSource{
								{
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: secretbox.GetSecretName(),
										},
									},
								},
								{
									ConfigMapRef: &corev1.ConfigMapEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: secretbox.GetConfigMapName(),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	ctrl.SetControllerReference(secretbox, deployment, r.Scheme)

	return deployment
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretboxReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&secretboxv1.Secretbox{}).
		Owns(&apps.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

// defaultAlphabet is the default alphabet used for generating random strings.
var defaultAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// generateRandomString generates a random string of the given length.
func generateRandomString(length int) (string, error) {
	if length < 1 {
		return "", fmt.Errorf("length must be greater than 0")
	}
	var generated string
	characterSet := strings.Split(defaultAlphabet, "")
	max := big.NewInt(int64(len(characterSet)))
	for i := 0; i < length; i++ {
		val, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		generated += characterSet[val.Int64()]
	}
	return generated, nil
}

// deletePods deletes all pods belonging to the given secretbox.
func (r *SecretboxReconciler) deletePods(ctx context.Context, secretbox *secretboxv1.Secretbox) error {
	log := log.FromContext(ctx)

	log.Info(fmt.Sprintf("Deleting all pods belonging to Secretbox %s", secretbox.Name))

	pods := &corev1.PodList{}
	if err := r.List(ctx, pods, &client.ListOptions{
		Namespace:     secretbox.Namespace,
		LabelSelector: labels.SelectorFromSet(secretbox.GetLabels()),
	}); err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	for _, pod := range pods.Items {
		if err := r.Delete(ctx, &pod); err != nil {
			return fmt.Errorf("failed to delete pod %s: %w", pod.Name, err)
		}
	}

	return nil
}
