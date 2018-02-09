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

package v1alpha1

import (
	v1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	scheme "github.com/giantswarm/apiextensions/pkg/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// AWSClusterConfigsGetter has a method to return a AWSClusterConfigInterface.
// A group's client should implement this interface.
type AWSClusterConfigsGetter interface {
	AWSClusterConfigs(namespace string) AWSClusterConfigInterface
}

// AWSClusterConfigInterface has methods to work with AWSClusterConfig resources.
type AWSClusterConfigInterface interface {
	Create(*v1alpha1.AWSClusterConfig) (*v1alpha1.AWSClusterConfig, error)
	Update(*v1alpha1.AWSClusterConfig) (*v1alpha1.AWSClusterConfig, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.AWSClusterConfig, error)
	List(opts v1.ListOptions) (*v1alpha1.AWSClusterConfigList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.AWSClusterConfig, err error)
	AWSClusterConfigExpansion
}

// aWSClusterConfigs implements AWSClusterConfigInterface
type aWSClusterConfigs struct {
	client rest.Interface
	ns     string
}

// newAWSClusterConfigs returns a AWSClusterConfigs
func newAWSClusterConfigs(c *CoreV1alpha1Client, namespace string) *aWSClusterConfigs {
	return &aWSClusterConfigs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the aWSClusterConfig, and returns the corresponding aWSClusterConfig object, and an error if there is any.
func (c *aWSClusterConfigs) Get(name string, options v1.GetOptions) (result *v1alpha1.AWSClusterConfig, err error) {
	result = &v1alpha1.AWSClusterConfig{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("awsclusterconfigs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of AWSClusterConfigs that match those selectors.
func (c *aWSClusterConfigs) List(opts v1.ListOptions) (result *v1alpha1.AWSClusterConfigList, err error) {
	result = &v1alpha1.AWSClusterConfigList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("awsclusterconfigs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested aWSClusterConfigs.
func (c *aWSClusterConfigs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("awsclusterconfigs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a aWSClusterConfig and creates it.  Returns the server's representation of the aWSClusterConfig, and an error, if there is any.
func (c *aWSClusterConfigs) Create(aWSClusterConfig *v1alpha1.AWSClusterConfig) (result *v1alpha1.AWSClusterConfig, err error) {
	result = &v1alpha1.AWSClusterConfig{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("awsclusterconfigs").
		Body(aWSClusterConfig).
		Do().
		Into(result)
	return
}

// Update takes the representation of a aWSClusterConfig and updates it. Returns the server's representation of the aWSClusterConfig, and an error, if there is any.
func (c *aWSClusterConfigs) Update(aWSClusterConfig *v1alpha1.AWSClusterConfig) (result *v1alpha1.AWSClusterConfig, err error) {
	result = &v1alpha1.AWSClusterConfig{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("awsclusterconfigs").
		Name(aWSClusterConfig.Name).
		Body(aWSClusterConfig).
		Do().
		Into(result)
	return
}

// Delete takes name of the aWSClusterConfig and deletes it. Returns an error if one occurs.
func (c *aWSClusterConfigs) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("awsclusterconfigs").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *aWSClusterConfigs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("awsclusterconfigs").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched aWSClusterConfig.
func (c *aWSClusterConfigs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.AWSClusterConfig, err error) {
	result = &v1alpha1.AWSClusterConfig{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("awsclusterconfigs").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
