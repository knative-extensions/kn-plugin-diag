package main

import (
	"fmt"

	. "github.com/cdlliuy/knative-diagnose/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	//"k8s.io/client-go/kubernetes"
	//"k8s.io/client-go/rest"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	networkingv1api "knative.dev/networking/pkg/apis/networking/v1alpha1"
	networkingv1alpha1 "knative.dev/networking/pkg/client/clientset/versioned/typed/networking/v1alpha1"

	servingv1api "knative.dev/serving/pkg/apis/serving/v1"
	v1 "knative.dev/serving/pkg/apis/serving/v1"
	servingv1client "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1"
)

func resources() {

	p := &ConnectionConfig{}
	p.Initialize()

	cfg, _ := p.RestConfig()

	servingClient, _ := servingv1client.NewForConfig(cfg)
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		fmt.Printf("failed to create k8s client%s\n", err)
	}

	nwclient, err := networkingv1alpha1.NewForConfig(cfg)
	if err != nil {
		fmt.Printf("failed to create serving client%s\n", err)
	}

	ksvcNsName := "default"
	ksvcName := "helloying"
	ksvc, err := servingClient.Services(ksvcNsName).Get(ksvcName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("failed to list service under namespace %s error:%v", err)
	} else {
		fmt.Printf("ksvc:%#v\n", ksvc.Spec.GetTemplate().Spec)
	}

	svcCreatedTime := ksvc.GetCreationTimestamp().Rfc3339Copy()
	svcConfigurationsReady := ksvc.Status.GetCondition(servingv1api.ServiceConditionConfigurationsReady).LastTransitionTime.Inner.Rfc3339Copy()
	svcRoutesReady := ksvc.Status.GetCondition(servingv1api.ServiceConditionRoutesReady).LastTransitionTime.Inner.Rfc3339Copy()

	svcConfigurationsReadyDuration := svcConfigurationsReady.Sub(svcCreatedTime.Time)
	svcRoutesReadyDuration := svcRoutesReady.Sub(svcCreatedTime.Time)
	svcReadyDuration := svcRoutesReady.Sub(svcCreatedTime.Time)

	cfgIns, err := servingClient.Configurations(ksvcNsName).Get(ksvcName, metav1.GetOptions{})
	if err != nil {
		fmt.Errorf("failed to get Configuration and skip measuring %s\n", err)
	}
	revisionName := cfgIns.Status.LatestReadyRevisionName

	revisionIns, err := servingClient.Revisions(ksvcNsName).Get(revisionName, metav1.GetOptions{})
	if err != nil {
		fmt.Errorf("failed to get Revision and skip measuring %s\n", err)
	}

	revisionCreatedTime := revisionIns.GetCreationTimestamp().Rfc3339Copy()
	revisionReadyTime := revisionIns.Status.GetCondition(v1.RevisionConditionReady).LastTransitionTime.Inner.Rfc3339Copy()
	revisionReadyDuration := revisionReadyTime.Sub(revisionCreatedTime.Time)

	label := fmt.Sprintf("serving.knative.dev/revision=%s", revisionName)
	podList := &corev1.PodList{}
	if podList, err = client.CoreV1().Pods(ksvcNsName).List(metav1.ListOptions{LabelSelector: label}); err != nil {
		fmt.Errorf("list Pods of revision[%s] error :%v", revisionName, err)
	}

	deploymentName := revisionName + "-deployment"
	deploymentIns, err := client.AppsV1().Deployments(ksvcNsName).Get(deploymentName, metav1.GetOptions{})
	if err != nil {
		fmt.Errorf("failed to find deployment of revision[%s] error:%v", revisionName, err)
	}

	deploymentCreatedTime := deploymentIns.GetCreationTimestamp().Rfc3339Copy()
	deploymentCreatedDuration := deploymentCreatedTime.Sub(revisionCreatedTime.Time)

	if len(podList.Items) > 0 {
		pod := podList.Items[0]
		podCreatedTime := pod.GetCreationTimestamp().Rfc3339Copy()
		present, PodScheduledCdt := podutil.GetPodCondition(&pod.Status, corev1.PodScheduled)
		if present == -1 {
			fmt.Errorf("failed to find Pod Condition PodScheduled and skip measuring")
		}
		podScheduledTime := PodScheduledCdt.LastTransitionTime.Rfc3339Copy()
		present, containersReadyCdt := podutil.GetPodCondition(&pod.Status, corev1.ContainersReady)
		if present == -1 {
			fmt.Errorf("failed to find Pod Condition ContainersReady and skip measuring")
		}
		containersReadyTime := containersReadyCdt.LastTransitionTime.Rfc3339Copy()
		podScheduledDuration := podScheduledTime.Sub(podCreatedTime.Time)
		containersReadyDuration := containersReadyTime.Sub(podCreatedTime.Time)

		queueProxyStatus, found := podutil.GetContainerStatus(pod.Status.ContainerStatuses, "queue-proxy")
		if !found {
			fmt.Errorf("failed to get queue-proxy container status and skip, error:%v", err)
		}
		queueProxyStartedTime := queueProxyStatus.State.Running.StartedAt.Rfc3339Copy()

		userContrainerStatus, found := podutil.GetContainerStatus(pod.Status.ContainerStatuses, "user-container")
		if !found {
			fmt.Errorf("failed to get user-container container status and skip, error:%v", err)
		}
		userContrainerStartedTime := userContrainerStatus.State.Running.StartedAt.Rfc3339Copy()

		queueProxyStartedDuration := queueProxyStartedTime.Sub(podCreatedTime.Time)
		userContrainerStartedDuration := userContrainerStartedTime.Sub(podCreatedTime.Time)

		fmt.Println(podScheduledDuration)
		fmt.Println(containersReadyDuration)
		fmt.Println(queueProxyStartedDuration)
		fmt.Println(userContrainerStartedDuration)
	}
	// TODO: Need to figure out a better way to measure PA time as its status keeps changing even after service creation.

	ingressIns, err := nwclient.Ingresses(ksvcNsName).Get(ksvcName, metav1.GetOptions{})
	if err != nil {
		fmt.Errorf("failed to get Ingress %s\n", err)
	}

	ingressCreatedTime := ingressIns.GetCreationTimestamp().Rfc3339Copy()
	ingressNetworkConfiguredTime := ingressIns.Status.GetCondition(networkingv1api.IngressConditionNetworkConfigured).LastTransitionTime.Inner.Rfc3339Copy()
	ingressLoadBalancerReadyTime := ingressIns.Status.GetCondition(networkingv1api.IngressConditionLoadBalancerReady).LastTransitionTime.Inner.Rfc3339Copy()
	ingressNetworkConfiguredDuration := ingressNetworkConfiguredTime.Sub(ingressCreatedTime.Time)
	ingressLoadBalancerReadyDuration := ingressLoadBalancerReadyTime.Sub(ingressNetworkConfiguredTime.Time)
	ingressReadyDuration := ingressLoadBalancerReadyTime.Sub(ingressCreatedTime.Time)

	fmt.Println(svcConfigurationsReadyDuration)
	fmt.Println(svcRoutesReadyDuration)
	fmt.Println(svcReadyDuration)
	fmt.Println(revisionReadyDuration)
	fmt.Println(deploymentCreatedDuration)

	fmt.Println(ingressNetworkConfiguredDuration)
	fmt.Println(ingressLoadBalancerReadyDuration)
	fmt.Println(ingressReadyDuration)

}
