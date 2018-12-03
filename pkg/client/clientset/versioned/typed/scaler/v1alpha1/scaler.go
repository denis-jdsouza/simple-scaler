/*
Copyright 2018 The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/arjunrn/simple-scaler/pkg/apis/scaler/v1alpha1"
	scheme "github.com/arjunrn/simple-scaler/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ScalersGetter has a method to return a ScalerInterface.
// A group's client should implement this interface.
type ScalersGetter interface {
	Scalers(namespace string) ScalerInterface
}

// ScalerInterface has methods to work with Scaler resources.
type ScalerInterface interface {
	Create(*v1alpha1.Scaler) (*v1alpha1.Scaler, error)
	Update(*v1alpha1.Scaler) (*v1alpha1.Scaler, error)
	UpdateStatus(*v1alpha1.Scaler) (*v1alpha1.Scaler, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Scaler, error)
	List(opts v1.ListOptions) (*v1alpha1.ScalerList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Scaler, err error)
	ScalerExpansion
}

// scalers implements ScalerInterface
type scalers struct {
	client rest.Interface
	ns     string
}

// newScalers returns a Scalers
func newScalers(c *ArjunnaikV1alpha1Client, namespace string) *scalers {
	return &scalers{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the scaler, and returns the corresponding scaler object, and an error if there is any.
func (c *scalers) Get(name string, options v1.GetOptions) (result *v1alpha1.Scaler, err error) {
	result = &v1alpha1.Scaler{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("scalers").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Scalers that match those selectors.
func (c *scalers) List(opts v1.ListOptions) (result *v1alpha1.ScalerList, err error) {
	result = &v1alpha1.ScalerList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("scalers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested scalers.
func (c *scalers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("scalers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a scaler and creates it.  Returns the server's representation of the scaler, and an error, if there is any.
func (c *scalers) Create(scaler *v1alpha1.Scaler) (result *v1alpha1.Scaler, err error) {
	result = &v1alpha1.Scaler{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("scalers").
		Body(scaler).
		Do().
		Into(result)
	return
}

// Update takes the representation of a scaler and updates it. Returns the server's representation of the scaler, and an error, if there is any.
func (c *scalers) Update(scaler *v1alpha1.Scaler) (result *v1alpha1.Scaler, err error) {
	result = &v1alpha1.Scaler{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("scalers").
		Name(scaler.Name).
		Body(scaler).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *scalers) UpdateStatus(scaler *v1alpha1.Scaler) (result *v1alpha1.Scaler, err error) {
	result = &v1alpha1.Scaler{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("scalers").
		Name(scaler.Name).
		SubResource("status").
		Body(scaler).
		Do().
		Into(result)
	return
}

// Delete takes name of the scaler and deletes it. Returns an error if one occurs.
func (c *scalers) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("scalers").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *scalers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("scalers").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched scaler.
func (c *scalers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Scaler, err error) {
	result = &v1alpha1.Scaler{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("scalers").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
