package utils

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	k8sv1api "k8s.io/api/apps/v1"
	k8scorev1api "k8s.io/api/core/v1"
	cachingv1alpha1api "knative.dev/caching/pkg/apis/caching/v1alpha1"
	nv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
	autoscalingv1v1alpha1api "knative.dev/serving/pkg/apis/autoscaling/v1alpha1"
	servingv1api "knative.dev/serving/pkg/apis/serving/v1"

	cachingv1alpha1 "knative.dev/caching/pkg/client/clientset/versioned/typed/caching/v1alpha1"
	networkingv1alpha1 "knative.dev/networking/pkg/client/clientset/versioned/typed/networking/v1alpha1"
	autoscalingv1alpha1 "knative.dev/serving/pkg/client/clientset/versioned/typed/autoscaling/v1alpha1"
	servingv1 "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1"
)

type Fetcher struct {
	ksvcName          string
	ksvcNSName        string
	conConfig         *ConnectionConfig
	KSVC              *servingv1api.Service
	Configuration     *servingv1api.Configuration
	Revision          *servingv1api.Revision
	Deployment        *k8sv1api.Deployment
	Images            *cachingv1alpha1api.Image
	KPA               *autoscalingv1v1alpha1api.PodAutoscaler
	Metrics           *autoscalingv1v1alpha1api.Metric
	SKS               *nv1alpha1.ServerlessService
	SKSPublicSVC      *k8scorev1api.Service
	SKSPrivateSVC     *k8scorev1api.Service
	SKSPublicEndpoint *k8scorev1api.Endpoints
	PrivateEendpoint  *k8scorev1api.Endpoints
	Routes            *servingv1api.Route
	RoutesSVC         *k8scorev1api.Service
	King              *nv1alpha1.Ingress
}

func NewFetcher(ksvcName, ksvcNSName string, p *ConnectionConfig) *Fetcher {
	return &Fetcher{
		ksvcName:   ksvcName,
		ksvcNSName: ksvcNSName,
		conConfig:  p,
	}
}

func (f *Fetcher) GetKSVCResources() error {
	//var deferErr error
	if f.conConfig == nil {
		panic("Missing connection config to contact with k8s cluster")
	}
	configuration, _ := f.conConfig.RestConfig()
	servingClient, err := servingv1.NewForConfig(configuration)
	if err != nil {
		fmt.Printf("failed to create kantive serving client%s\n", err)
	}
	k8sClient, err := kubernetes.NewForConfig(configuration)
	if err != nil {
		fmt.Printf("failed to create k8s client%s\n", err)
	}
	cachingClient, err := cachingv1alpha1.NewForConfig(configuration)
	if err != nil {
		fmt.Printf("failed to create knative caching client%s\n", err)
	}
	autoscalingClient, err := autoscalingv1alpha1.NewForConfig(configuration)
	if err != nil {
		fmt.Printf("failed to create knative autoscaling client%s\n", err)
	}

	networkingClient, err := networkingv1alpha1.NewForConfig(configuration)
	if err != nil {
		fmt.Printf("failed to create knative networking client%s\n", err)
	}

	ksvc, err := servingClient.Services(f.ksvcNSName).Get(f.ksvcName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get Knative Service under namespace %s !  Error: %v", f.ksvcNSName, err)
	}
	f.KSVC = ksvc

	cfg, err := servingClient.Configurations(f.ksvcNSName).Get(f.ksvcName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get Configuration for Knative Service %s under namespace %s ! Error: %v", f.ksvcName, f.ksvcNSName, err)
	}
	f.Configuration = cfg
	revisionName := cfg.Status.LatestCreatedRevisionName

	revision, err := servingClient.Revisions(f.ksvcNSName).Get(revisionName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get Knative Revision for Knative Service %s under namespace %s! Error: %v", f.ksvcName, f.ksvcNSName, err)
	}
	f.Revision = revision

	deployment, err := k8sClient.AppsV1().Deployments(f.ksvcNSName).Get(revisionName+"-deployment", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get K8S deployment for Knative Service %s / Revision %s under namespace %s! Error: %v", f.ksvcName, revisionName, f.ksvcNSName, err)
	}
	f.Deployment = deployment

	image, err := cachingClient.Images(f.ksvcNSName).Get(revisionName+"-cache-user-container", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get Knative image caching resources for Knative Service %s / Revision %s under namespace %s! Error: %v", f.ksvcName, revisionName, f.ksvcNSName, err)
	}
	f.Images = image

	kpa, err := autoscalingClient.PodAutoscalers(f.ksvcNSName).Get(revisionName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get Knative podautoscaler resources for Knative Service %s / Revision %s under namespace %s! Error: %v", f.ksvcName, revisionName, f.ksvcNSName, err)
	}
	f.KPA = kpa

	metrics, err := autoscalingClient.Metrics(f.ksvcNSName).Get(revisionName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get Knative metrics resources for Knative Service %s / Revision %s under namespace %s! Error: %v", f.ksvcName, revisionName, f.ksvcNSName, err)
	}
	f.Metrics = metrics

	sks, err := networkingClient.ServerlessServices(f.ksvcNSName).Get(revisionName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get Knative ServerlessService resources for Knative Service %s / Revision %s under namespace %s! Error: %v", f.ksvcName, revisionName, f.ksvcNSName, err)
	}
	f.SKS = sks

	publicsvc, err := k8sClient.CoreV1().Services(f.ksvcNSName).Get(revisionName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get K8S public service for Knative Service %s / Revision %s under namespace %s! Error: %v", f.ksvcName, revisionName, f.ksvcNSName, err)
	}
	f.SKSPublicSVC = publicsvc

	privatesvc, err := k8sClient.CoreV1().Services(f.ksvcNSName).Get(revisionName+"-private", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get K8S public service for Knative Service %s / Revision %s under namespace %s! Error: %v", f.ksvcName, revisionName, f.ksvcNSName, err)
	}
	f.SKSPrivateSVC = privatesvc

	publicep, err := k8sClient.CoreV1().Endpoints(f.ksvcNSName).Get(revisionName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get K8S public endpoint for Knative Service %s / Revision %s under namespace %s! Error: %v", f.ksvcName, revisionName, f.ksvcNSName, err)
	}
	f.SKSPublicEndpoint = publicep

	privateep, err := k8sClient.CoreV1().Endpoints(f.ksvcNSName).Get(revisionName+"-private", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get K8S public endpoint for Knative Service %s / Revision %s under namespace %s! Error: %v", f.ksvcName, revisionName, f.ksvcNSName, err)
	}
	f.PrivateEendpoint = privateep

	routes, err := servingClient.Routes(f.ksvcNSName).Get(f.ksvcName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get Knative Routes resources for Knative Service %s under namespace %s! Error: %v", f.ksvcName, f.ksvcNSName, err)
	}
	f.Routes = routes

	externalsvc, err := k8sClient.CoreV1().Services(f.ksvcNSName).Get(f.ksvcName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get K8S external public service for Knative Service %s / Revision %s under namespace %s! Error: %v", f.ksvcName, revisionName, f.ksvcNSName, err)
	}
	f.RoutesSVC = externalsvc

	king, err := networkingClient.Ingresses(f.ksvcNSName).Get(f.ksvcName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get Knative Ingress resources for Knative Service %s under namespace %s! Error: %v", f.ksvcName, f.ksvcNSName, err)
	}
	f.King = king

	return nil

}
