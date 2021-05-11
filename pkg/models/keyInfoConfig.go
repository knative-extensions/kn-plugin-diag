package models

import (
	"encoding/json"
	"log"
)

type KeyInfoConfiguration struct {
	Name     string   `json:"name"`
	KeyInfos []string `json:"keyInfos"`
}

func defaultServingKeyInfoConfiguration() []byte {
	//support slice by [*] only
	configurationJSON := `
[
	{
		"name": "ksvc",
		"keyInfos": [
			"spec.template.metadata.annotations",
			"spec.template.metadata.name",
			"spec.template.spec.containerConcurrency",
			"spec.template.spec.enableServiceLinks",
			"spec.template.spec.timeoutSeconds",
			"spec.template.spec.containers[*].image",
			"spec.template.spec.containers[*].resources.requests",
			"spec.template.spec.containers[*].resources.limits",
			"status.latestCreatedRevisionName",
			"status.latestReadyRevisionName",
			"spec.traffic[*]",
			"status.traffic[*]",
			"status.address",
			"status.url"
		]
	},
	{
		"name": "configuration",
		"keyInfos": [
		]
	},
	{
		"name": "revision",
		"keyInfos": [
		]
	},
	{
		"name": "image",
		"keyInfos": [
			"spec.image"
		]
	},
	{
		"name": "deployment",
		"keyInfos": [
			"spec.progressDeadlineSeconds",
			"spec.replicas",
			"status.availableReplicas",
			"status.readyReplicas",
			"spec.template.spec.containers[*].image"
		]
	},
	{
		"name": "replicaset",
		"keyInfos": [
		]
	},
	{
		"name": "pod",
		"keyInfos": [
			"spec.tolerations[*]",
			"spec.nodeName",
			"spec.containers[*].resources.requests",
			"spec.containers[*].resources.limits",
			"spec.containers[*].imagePullPolicy",
			"status.phase",
			"status.containerStatuses[*].restartCount",
			"status.containerStatuses[*].state",
			"status.containerStatuses[*].lastState"
		]
	},
	{
		"name": "kpa",
		"keyInfos": [
			"spec.scaleTargetRef",
			"status.desiredScale",
			"status.actualScale",
			"status.metricsServiceName",
			"status.serviceName"
		]
	},
	{
		"name": "metric",
		"keyInfos": [
			"spec.panicWindow",
			"spec.stableWindow",
			"spec.scrapeTarget"
		]
	},
	{
		"name": "sks",
		"keyInfos": [
			"spec.mode",
			"spec.numActivators",
			"status.privateServiceName",
			"status.serviceName"
		]
	},
	{
		"name": "publicSVC",
		"keyInfos": [
			"spec.clusterIPs[*]",
			"spec.ports[*]",
			"spec.type"
		]
	},
	{
		"name": "publicEndpoint",
		"keyInfos": [
			"subsets[*].addresses[*].ip",
			"subsets[*].ports"
		]
	},
	{
		"name": "privateSVC",
		"keyInfos": [
			"spec.clusterIPs[*]",
			"spec.ports[*]",
			"spec.type"
		]
	},
	{
		"name": "privateEndpoint",
		"keyInfos": [
			"subsets[*].addresses[*].ip",
			"subsets[*].ports"
		]
	},
	{
		"name": "route",
		"keyInfos": [
		]
	},
	{
		"name": "configuration",
		"keyInfos": [
			"spec.traffic"
		]
	},
	{
		"name": "externalSVC",
		"keyInfos": [
			"spec.externalName",
			"spec.ports[*]",
			"spec.type"
		]
	},
	{
		"name": "kingress",
		"keyInfos": [
			"spec.rules[*].hosts",
			"spec.rules[*].http.paths[*]",
			"spec.rules[*].visibility",
			"status.privateLoadBalancer.ingress[*]",
			"status.publicLoadBalancer.ingress[*]"
		]
	}	
]`

	return []byte(configurationJSON)
}

func LoadServingKeyInfoConfiguration() map[string][]string {

	var configurations []KeyInfoConfiguration
	if err := json.Unmarshal(defaultServingKeyInfoConfiguration(), &configurations); err != nil {
		log.Fatal(err)
	}

	keyinfoMap := make(map[string][]string)
	for _, item := range configurations {
		if _, ok := keyinfoMap[item.Name]; !ok {
			keyinfoMap[item.Name] = item.KeyInfos
		}
	}
	return keyinfoMap
}
