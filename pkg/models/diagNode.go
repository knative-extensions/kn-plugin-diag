/*
Copyright 2021 The Knative Authors

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

package models

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type CRNode struct {
	Name            string
	GVR             schema.GroupVersionResource
	GetResourceName func(string) string
	GetListOptions  func([]string) metav1.ListOptions
	Leaves          []*CRNode
}

type ObjectNode struct {
	CRName     string
	ObjectName string
	Object     *unstructured.Unstructured
	Leaves     []*ObjectNode
}

func NewCRNode(name string, gvr schema.GroupVersionResource, f ...func(string) string) *CRNode {
	t := &CRNode{
		Name: name,
		GVR:  gvr,
	}
	if len(f) != 0 {
		t.GetResourceName = f[0]
		t.GetListOptions = nil
	}
	return t
}

func (t *CRNode) SetListOptions(f func([]string) metav1.ListOptions) {
	t.GetListOptions = f
	t.GetResourceName = nil
}

func (t *CRNode) AddLeafNode(leaf *CRNode) {
	t.Leaves = append(t.Leaves, leaf)
}

func NewObjectNode(crName, objectName string, object *unstructured.Unstructured) *ObjectNode {
	tr := &ObjectNode{
		CRName:     crName,
		ObjectName: objectName,
		Object:     object,
	}
	return tr
}

func (t *ObjectNode) AddLeafNode(leaf *ObjectNode) {
	t.Leaves = append(t.Leaves, leaf)
}
