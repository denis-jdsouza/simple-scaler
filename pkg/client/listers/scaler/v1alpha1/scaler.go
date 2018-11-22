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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/arjunrn/dumb-scaler/pkg/apis/scaler/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ScalerLister helps list Scalers.
type ScalerLister interface {
	// List lists all Scalers in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.Scaler, err error)
	// Scalers returns an object that can list and get Scalers.
	Scalers(namespace string) ScalerNamespaceLister
	ScalerListerExpansion
}

// scalerLister implements the ScalerLister interface.
type scalerLister struct {
	indexer cache.Indexer
}

// NewScalerLister returns a new ScalerLister.
func NewScalerLister(indexer cache.Indexer) ScalerLister {
	return &scalerLister{indexer: indexer}
}

// List lists all Scalers in the indexer.
func (s *scalerLister) List(selector labels.Selector) (ret []*v1alpha1.Scaler, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Scaler))
	})
	return ret, err
}

// Scalers returns an object that can list and get Scalers.
func (s *scalerLister) Scalers(namespace string) ScalerNamespaceLister {
	return scalerNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ScalerNamespaceLister helps list and get Scalers.
type ScalerNamespaceLister interface {
	// List lists all Scalers in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.Scaler, err error)
	// Get retrieves the Scaler from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.Scaler, error)
	ScalerNamespaceListerExpansion
}

// scalerNamespaceLister implements the ScalerNamespaceLister
// interface.
type scalerNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Scalers in the indexer for a given namespace.
func (s scalerNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Scaler, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Scaler))
	})
	return ret, err
}

// Get retrieves the Scaler from the indexer for a given namespace and name.
func (s scalerNamespaceLister) Get(name string) (*v1alpha1.Scaler, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("scaler"), name)
	}
	return obj.(*v1alpha1.Scaler), nil
}
