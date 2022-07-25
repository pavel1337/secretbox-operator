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
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	secretboxv1 "github.com/pavel1337/secretbox-operator/api/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func Test_generateRandomString(t *testing.T) {
	_, err := generateRandomString(0)
	assert.Error(t, err)

	_, err = generateRandomString(-1)
	assert.Error(t, err)

	for i := 1; i < 100; i++ {
		s, err := generateRandomString(i)
		assert.NoError(t, err)
		assert.Equal(t, i, len(s))
	}

	for i := 0; i < 1000; i++ {
		one, err := generateRandomString(32)
		assert.NoError(t, err)
		two, err := generateRandomString(32)
		assert.NoError(t, err)
		assert.NotEqual(t, one, two)
	}
}

var _ = Describe("SecretboxController", func() {
	const (
		namespace = "default"
		name      = "sample"
	)

	var secretboxLookupKey = types.NamespacedName{Name: name, Namespace: namespace}

	Context("when creating a Secretbox", func() {
		It("should not create a Secretbox with invalid size", func() {
			ctx := context.Background()
			secretbox := &secretboxv1.Secretbox{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: secretboxv1.SecretboxSpec{
					Size: 0,
				},
			}

			Expect(k8sClient.Create(ctx, secretbox)).To(HaveOccurred())
		})

		It("should create a Secretbox with downstream objects", func() {
			ctx := context.Background()
			secretbox := &secretboxv1.Secretbox{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: secretboxv1.SecretboxSpec{
					Size: 1,
				},
			}

			Expect(k8sClient.Create(ctx, secretbox)).To(Succeed())

			createdSecretbox := &secretboxv1.Secretbox{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, secretboxLookupKey, createdSecretbox)
				if err != nil {
					return false
				}
				return true
			}, time.Minute, time.Second).Should(BeTrue())
			Expect(createdSecretbox.Name).To(Equal(name))

			Eventually(func() bool {
				// get the secret object
				secret := &corev1.Secret{}
				err := k8sClient.Get(ctx, types.NamespacedName{Name: createdSecretbox.GetSecretName(), Namespace: namespace}, secret)
				if err != nil {
					return false
				}

				// get the configmap object
				configmap := &corev1.ConfigMap{}
				err = k8sClient.Get(ctx, types.NamespacedName{Name: createdSecretbox.GetConfigMapName(), Namespace: namespace}, configmap)
				if err != nil {
					return false
				}

				// get the deployment object
				deployment := &appsv1.Deployment{}
				err = k8sClient.Get(ctx, types.NamespacedName{Name: createdSecretbox.GetDeploymentName(), Namespace: namespace}, deployment)
				if err != nil {
					return false
				}

				// get the service object
				service := &corev1.Service{}
				err = k8sClient.Get(ctx, types.NamespacedName{Name: createdSecretbox.GetServiceName(), Namespace: namespace}, service)
				if err != nil {
					return false
				}

				return true

			}, time.Minute, time.Second).Should(BeTrue())
		})

	})

	Context("when child service is delete", func() {
		It("should recreate service", func() {
			ctx := context.Background()
			secretbox := &secretboxv1.Secretbox{}
			err := k8sClient.Get(ctx, secretboxLookupKey, secretbox)
			Expect(err).To(Succeed())

			// get service owned by secretbox
			service := &corev1.Service{}
			serviceLookupKey := types.NamespacedName{Name: secretbox.GetServiceName(), Namespace: secretbox.Namespace}
			err = k8sClient.Get(ctx, serviceLookupKey, service)
			Expect(err).To(Succeed())
			Expect(service.Name).To(Equal(secretbox.GetServiceName()))

			// delete service
			err = k8sClient.Delete(ctx, service)
			Expect(err).To(Succeed())

			// wait for service to be recreated
			Eventually(func() bool {
				err := k8sClient.Get(ctx, serviceLookupKey, service)
				if err != nil {
					return false
				}
				return true
			}, time.Minute, time.Second).Should(BeTrue())
			Expect(service.Name).To(Equal(secretbox.GetServiceName()))
		})
	})
})
