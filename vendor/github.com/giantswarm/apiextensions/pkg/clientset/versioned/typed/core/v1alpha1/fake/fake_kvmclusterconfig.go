/*
Copyright 2018 Giant Swarm GmbH.

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

package fake

import (
	v1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeKVMClusterConfigs implements KVMClusterConfigInterface
type FakeKVMClusterConfigs struct {
	Fake *FakeCoreV1alpha1
	ns   string
}

var kvmclusterconfigsResource = schema.GroupVersionResource{Group: "core.giantswarm.io", Version: "v1alpha1", Resource: "kvmclusterconfigs"}

var kvmclusterconfigsKind = schema.GroupVersionKind{Group: "core.giantswarm.io", Version: "v1alpha1", Kind: "KVMClusterConfig"}

// Get takes name of the kVMClusterConfig, and returns the corresponding kVMClusterConfig object, and an error if there is any.
func (c *FakeKVMClusterConfigs) Get(name string, options v1.GetOptions) (result *v1alpha1.KVMClusterConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(kvmclusterconfigsResource, c.ns, name), &v1alpha1.KVMClusterConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.KVMClusterConfig), err
}

// List takes label and field selectors, and returns the list of KVMClusterConfigs that match those selectors.
func (c *FakeKVMClusterConfigs) List(opts v1.ListOptions) (result *v1alpha1.KVMClusterConfigList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(kvmclusterconfigsResource, kvmclusterconfigsKind, c.ns, opts), &v1alpha1.KVMClusterConfigList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.KVMClusterConfigList{}
	for _, item := range obj.(*v1alpha1.KVMClusterConfigList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested kVMClusterConfigs.
func (c *FakeKVMClusterConfigs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(kvmclusterconfigsResource, c.ns, opts))

}

// Create takes the representation of a kVMClusterConfig and creates it.  Returns the server's representation of the kVMClusterConfig, and an error, if there is any.
func (c *FakeKVMClusterConfigs) Create(kVMClusterConfig *v1alpha1.KVMClusterConfig) (result *v1alpha1.KVMClusterConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(kvmclusterconfigsResource, c.ns, kVMClusterConfig), &v1alpha1.KVMClusterConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.KVMClusterConfig), err
}

// Update takes the representation of a kVMClusterConfig and updates it. Returns the server's representation of the kVMClusterConfig, and an error, if there is any.
func (c *FakeKVMClusterConfigs) Update(kVMClusterConfig *v1alpha1.KVMClusterConfig) (result *v1alpha1.KVMClusterConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(kvmclusterconfigsResource, c.ns, kVMClusterConfig), &v1alpha1.KVMClusterConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.KVMClusterConfig), err
}

// Delete takes name of the kVMClusterConfig and deletes it. Returns an error if one occurs.
func (c *FakeKVMClusterConfigs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(kvmclusterconfigsResource, c.ns, name), &v1alpha1.KVMClusterConfig{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeKVMClusterConfigs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(kvmclusterconfigsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.KVMClusterConfigList{})
	return err
}

// Patch applies the patch and returns the patched kVMClusterConfig.
func (c *FakeKVMClusterConfigs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.KVMClusterConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(kvmclusterconfigsResource, c.ns, name, data, subresources...), &v1alpha1.KVMClusterConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.KVMClusterConfig), err
}
