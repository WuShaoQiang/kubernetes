/*
Copyright 2017 The Kubernetes Authors.

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

package configuration

import (
	"reflect"
	"testing"

	"k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetValidatingWebhookConfig(t *testing.T) {
	// Build a test client that the admission plugin can use to look up the ValidatingWebhookConfiguration
	client := fake.NewSimpleClientset()
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	stop := make(chan struct{})
	defer close(stop)
	informerFactory.Start(stop)
	informerFactory.WaitForCacheSync(stop)

	manager := NewValidatingWebhookConfigurationManager(informerFactory)
	if validatingConfig, ok := manager.(*validatingWebhookConfigurationManager); ok {
		validatingConfig.updateConfiguration()
	}
	// no configurations
	if configurations := manager.Webhooks(); len(configurations) != 0 {
		t.Errorf("expected empty webhooks, but got %v", configurations)
	}

	webhookConfiguration := &v1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "webhook1"},
		Webhooks:   []v1.ValidatingWebhook{{Name: "webhook1.1"}},
	}

	validatingInformer := informerFactory.Admissionregistration().V1().ValidatingWebhookConfigurations()
	validatingInformer.Informer().GetIndexer().Add(webhookConfiguration)
	if validatingConfig, ok := manager.(*validatingWebhookConfigurationManager); ok {
		validatingConfig.updateConfiguration()
	}
	// configuration populated
	configurations := manager.Webhooks()
	if len(configurations) == 0 {
		t.Errorf("expected non empty webhooks")
	}
	for i := range configurations {
		h, ok := configurations[i].GetValidatingWebhook()
		if !ok {
			t.Errorf("Expected validating webhook")
			continue
		}
		if !reflect.DeepEqual(h, &webhookConfiguration.Webhooks[i]) {
			t.Errorf("Expected\n%#v\ngot\n%#v", &webhookConfiguration.Webhooks[i], h)
		}
	}
}
