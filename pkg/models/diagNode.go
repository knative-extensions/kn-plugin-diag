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
