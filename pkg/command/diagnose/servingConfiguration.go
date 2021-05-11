package diagnose

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	. "knative.dev/kn-plugin-diag/pkg/models"
	"knative.dev/kn-plugin-diag/pkg/utils"
	. "knative.dev/kn-plugin-diag/pkg/utils"
	"knative.dev/serving/pkg/apis/serving"
)

type ServingConfiguration struct {
	ksvcName                string
	Namespace               string
	lastCreatedRevisionName string
	dynClient               dynamic.Interface
	crdRoot                 *CRNode
	objectRoot              *ObjectNode
	keyInfos                map[string][]string
	conditionInfos          map[string][]ConditionInfo
}

func NewServingConfiguration(ksvcName, Namespace string, p *ConnectionConfig) (*ServingConfiguration, error) {

	if p == nil {
		return nil, fmt.Errorf("Missing connection config to contact with k8s cluster. Please set context of KUBECONFIG")
	}
	configuration, _ := p.RestConfig()

	dynClient, err := dynamic.NewForConfig(configuration)
	if err != nil {
		return nil, fmt.Errorf("Failed to create dynamic client %v\n", err)
	}

	sc := &ServingConfiguration{
		ksvcName:  ksvcName,
		Namespace: Namespace,
		dynClient: dynClient,
	}

	sc.initCRDHierarchy()
	sc.addKeyInfo()
	sc.addConditionInfo()
	LoadServingConditionInfoConfiguration()
	err = sc.buildObjectHierarchy(sc.crdRoot)
	if err != nil {
		return nil, err
	}
	return sc, nil
}

func (sc *ServingConfiguration) initCRDHierarchy() {
	ksvc := NewCRNode("ksvc", schema.GroupVersionResource{
		Group:    "serving.knative.dev",
		Version:  "v1",
		Resource: "services"},
		func(ksvcName string) string {
			return ksvcName
		})
	configuration := NewCRNode("configuration", schema.GroupVersionResource{
		Group:    "serving.knative.dev",
		Version:  "v1",
		Resource: "configurations",
	}, func(ksvcName string) string {
		return ksvcName
	})
	route := NewCRNode("route", schema.GroupVersionResource{
		Group:    "serving.knative.dev",
		Version:  "v1",
		Resource: "routes",
	}, func(ksvcName string) string {
		return ksvcName
	})
	revision := NewCRNode("revision", schema.GroupVersionResource{
		Group:    "serving.knative.dev",
		Version:  "v1",
		Resource: "revisions",
	}, func(lastCreatedRevisionName string) string {
		return lastCreatedRevisionName
	})

	image := NewCRNode("image", schema.GroupVersionResource{
		Group:    "caching.internal.knative.dev",
		Version:  "v1alpha1",
		Resource: "images",
	}, func(lastCreatedRevisionName string) string {
		return lastCreatedRevisionName + "-cache-user-container"
	})

	deployment := NewCRNode("deployment", schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}, func(lastCreatedRevisionName string) string {
		return lastCreatedRevisionName + "-deployment"
	})

	replicaset := NewCRNode("replicaset", schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "replicasets",
	})
	replicaset.SetListOptions(func(labels []string) metav1.ListOptions {
		return metav1.ListOptions{
			LabelSelector: strings.Join(labels, ","),
		}
	})

	pod := NewCRNode("pod", schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	})
	pod.SetListOptions(func(labels []string) metav1.ListOptions {
		return metav1.ListOptions{
			LabelSelector: strings.Join(labels, ","),
			Limit:         5,
		}
	})

	kpa := NewCRNode("kpa", schema.GroupVersionResource{
		Group:    "autoscaling.internal.knative.dev",
		Version:  "v1alpha1",
		Resource: "podautoscalers",
	}, func(lastCreatedRevisionName string) string {
		return lastCreatedRevisionName
	})

	metric := NewCRNode("metric", schema.GroupVersionResource{
		Group:    "autoscaling.internal.knative.dev",
		Version:  "v1alpha1",
		Resource: "metrics",
	}, func(lastCreatedRevisionName string) string {
		return lastCreatedRevisionName
	})

	sks := NewCRNode("sks", schema.GroupVersionResource{
		Group:    "networking.internal.knative.dev",
		Version:  "v1alpha1",
		Resource: "serverlessservices",
	}, func(lastCreatedRevisionName string) string {
		return lastCreatedRevisionName
	})

	publicSVC := NewCRNode("publicSVC", schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "services",
	}, func(lastCreatedRevisionName string) string {
		return lastCreatedRevisionName
	})

	privateSVC := NewCRNode("privateSVC", schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "services",
	}, func(lastCreatedRevisionName string) string {
		return lastCreatedRevisionName + "-private"
	})

	publicEndpoint := NewCRNode("publicEndpoint", schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "endpoints",
	}, func(lastCreatedRevisionName string) string {
		return lastCreatedRevisionName
	})

	privateEndpoint := NewCRNode("privateEndpoint", schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "endpoints",
	}, func(lastCreatedRevisionName string) string {
		return lastCreatedRevisionName + "-private"
	})

	externalSVC := NewCRNode("externalSVC", schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "services",
	}, func(ksvcName string) string {
		return ksvcName
	})

	ingress := NewCRNode("kingress", schema.GroupVersionResource{
		Group:    "networking.internal.knative.dev",
		Version:  "v1alpha1",
		Resource: "ingresses",
	}, func(ksvcName string) string {
		return ksvcName
	})

	ksvc.AddLeafNode(configuration)
	ksvc.AddLeafNode(route)
	configuration.AddLeafNode(revision)
	revision.AddLeafNode(image)
	revision.AddLeafNode(deployment)
	revision.AddLeafNode(kpa)
	deployment.AddLeafNode(replicaset)
	replicaset.AddLeafNode(pod)
	kpa.AddLeafNode(metric)
	kpa.AddLeafNode(sks)
	sks.AddLeafNode(publicSVC)
	sks.AddLeafNode(privateSVC)
	publicSVC.AddLeafNode(publicEndpoint)
	privateSVC.AddLeafNode(privateEndpoint)
	route.AddLeafNode(externalSVC)
	route.AddLeafNode(ingress)

	sc.crdRoot = ksvc

}

func (sc *ServingConfiguration) addKeyInfo() {
	sc.keyInfos = LoadServingKeyInfoConfiguration()
}

func (sc *ServingConfiguration) addConditionInfo() {
	sc.conditionInfos = LoadServingConditionInfoConfiguration()
}

func (sc *ServingConfiguration) buildObjectHierarchy(crNode *CRNode, parentObjectsNode ...*ObjectNode) error {

	if crNode == nil {
		return nil
	}

	objectNodes := []*ObjectNode{}

	if crNode.GetResourceName == nil && crNode.GetListOptions == nil {
		return fmt.Errorf("Invalid CRD definition %s, missing both GetResourceName and GetListOptions definition.", crNode.Name)
	}

	if crNode.GetResourceName != nil {
		objectName := ""
		switch crNode.Name {
		case "ksvc", "configuration", "route", "externalSVC", "kingress":
			objectName = crNode.GetResourceName(sc.ksvcName)
		default:
			objectName = crNode.GetResourceName(sc.lastCreatedRevisionName)
		}

		obj, err := sc.dynClient.Resource(crNode.GVR).Namespace(sc.Namespace).Get(context.Background(), objectName, metav1.GetOptions{})
		if err != nil {
			utils.SayWarningMessage("Failed to load resource %s of %s,  %v\n", crNode.Name, objectName, err)
			return nil
			//return fmt.Errorf("Failed to load resource %s of %s,  %v\n", crNode.Name, objectName, err)
		}

		objectNode := NewObjectNode(crNode.Name, objectName, obj)
		objectNodes = append(objectNodes, objectNode)
		//link the current object to its owner object
		for _, parent := range parentObjectsNode {
			parent.Leaves = append(parent.Leaves, objectNode)
		}

		//special handling for ksvc to complete the initialization of `serviceconfiguration` struct
		if crNode.Name == "ksvc" {
			sc.objectRoot = objectNode
			lastCreatedRevisionName, ok, err := unstructured.NestedString(obj.Object, strings.Split("status.latestCreatedRevisionName", ".")...)
			if ok && err == nil {
				sc.lastCreatedRevisionName = lastCreatedRevisionName
			} else {
				utils.SayWarningMessage("Failed to load the lastCreatedRevisionName from %s of %s, %v\n", crNode.Name, sc.ksvcName, err)
				return nil
			}
		}
	}

	if crNode.GetListOptions != nil {
		//handle list options for revision and pods
		for _, parent := range parentObjectsNode {

			if crNode.Name == "replicaset" {
				listOptions := crNode.GetListOptions([]string{
					serving.RevisionLabelKey + "=" + sc.lastCreatedRevisionName,
				})

				objList, err := sc.dynClient.Resource(crNode.GVR).Namespace(sc.Namespace).List(context.Background(), listOptions)
				if err != nil {
					utils.SayWarningMessage("Failed to load resource %s with label %s, %v\n", crNode.Name, listOptions.LabelSelector, err)
					return nil
				}

				for _, obj := range objList.Items {
					objectName, ok, err := unstructured.NestedString(obj.Object, strings.Split("metadata.name", ".")...)
					if !ok || err != nil {
						utils.SayWarningMessage("Failed to load the metadata.name from %s of %s, %v\n", crNode.Name, sc.ksvcName, err)
						return nil
					}
					replicaNumer, ok, err := unstructured.NestedInt64(obj.Object, strings.Split("spec.replicas", ".")...)
					if !ok || err != nil {
						utils.SayWarningMessage("Failed to load the spec.replicas from %s of %s, %v\n", crNode.Name, sc.ksvcName, err)
						return nil
					}
					//only count the replicaset with desired number > 0
					if replicaNumer > 0 {
						objectNode := NewObjectNode(crNode.Name, objectName, &obj)
						objectNodes = append(objectNodes, objectNode)
						parent.Leaves = append(parent.Leaves, objectNode)
					}
				}
			} //end of if replicaset

			if crNode.Name == "pod" {
				podhash := strings.TrimPrefix(parent.ObjectName, sc.lastCreatedRevisionName+"-deployment-")
				listOptions := crNode.GetListOptions([]string{
					serving.RevisionLabelKey + "=" + sc.lastCreatedRevisionName,
					"pod-template-hash" + "=" + podhash,
				})

				objList, err := sc.dynClient.Resource(crNode.GVR).Namespace(sc.Namespace).List(context.Background(), listOptions)
				if err != nil {
					utils.SayWarningMessage("Failed to load resource %s with label %s, %v\n", crNode.Name, listOptions.LabelSelector, err)
					return nil
				}

				for _, obj := range objList.Items {
					objectName, ok, err := unstructured.NestedString(obj.Object, strings.Split("metadata.name", ".")...)
					if !ok || err != nil {
						utils.SayWarningMessage("Failed to load the metadata.name for %s %s, %v\n", crNode.Name, sc.ksvcName, err)
						return nil
					}
					objectNode := NewObjectNode(crNode.Name, objectName, &obj)
					objectNodes = append(objectNodes, objectNode)
					parent.Leaves = append(parent.Leaves, objectNode)
				}
			} //end of `if pod`

		} //end of `for parent`
	} //end of `if listOptions`

	for _, leaf := range crNode.Leaves {
		err := sc.buildObjectHierarchy(leaf, objectNodes...)
		if err != nil {
			return err
		}
	}

	return nil

}

func (sc *ServingConfiguration) deepFirstRetrieve(node *CRNode, depth int) {

	if node == nil {
		return
	}
	if depth == 0 {
		fmt.Printf("%s(0)\n", node.Name)
	} else {
		fmt.Printf("|%s %s(%d)\n", strings.Repeat("-", depth*2-1), node.Name, depth)
	}
	for _, leaf := range node.Leaves {
		sc.deepFirstRetrieve(leaf, depth+1)
	}

}

func (sc *ServingConfiguration) deepFirstRetrieveObjects(node *ObjectNode, depth int, table Table, verbose string) error {

	if node == nil {
		return nil
	}

	var printResource *PrintableResource
	var err error
	if verbose == "keyinfo" {
		printResource = NewPrintableResource(depth, node.CRName, node.ObjectName, WithVerboseType(verbose))

		//only apply to the CRs that have keyInfo definition in keyInfoConfiguration
		if keyInfo, ok := sc.keyInfos[node.CRName]; ok {
			err = printResource.AddKeyInfo(keyInfo, node)
		}
		if err != nil {
			return err
		}
	} else {
		creationTimestamp, ok, err := unstructured.NestedString(node.Object.Object, strings.Split("metadata.creationTimestamp", ".")...)
		if !ok || err != nil {
			utils.SayWarningMessage("Failed to load the metadata.creationTimestamp for %s %s, %v\n", node.CRName, node.ObjectName, err)
			return nil
		}
		printResource = NewPrintableResource(depth, node.CRName, node.ObjectName, WithCreatedAt(creationTimestamp), WithVerboseType(verbose))
		err = printResource.AddConditions(node, sc.conditionInfos)
		if err != nil {
			return err
		}
	}

	table.AddMuitpleRows(printResource.DumpResource())

	for _, leaf := range node.Leaves {
		err = sc.deepFirstRetrieveObjects(leaf, depth+1, table, verbose)
		if err != nil {
			return err
		}
	}
	return nil

}
