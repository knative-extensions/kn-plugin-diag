package diagnose

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	. "github.com/knative-sandbox/kn-plugin-diag/pkg/utils"
	"github.com/spf13/cobra"

	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	nv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
	autoscalingv1v1alpha1api "knative.dev/serving/pkg/apis/autoscaling/v1alpha1"
	servingv1api "knative.dev/serving/pkg/apis/serving/v1"
)

var (
	n       string
	verbose string
)

// domainCmd represents the domain command
func NewServiceCmd(p *ConnectionConfig) *cobra.Command {
	var serviceCmd = &cobra.Command{
		Use:   "service",
		Short: "kn-diag service",
		Long: `Query knative service details. For example
knative-diagnose service <ksvc-name>\n`,

		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf(`'service' requires a input arguments for knative servie name. For example: 
knative-diagnose service <ksvc-name> -ns <namespace>\n`)
			}
			//if cmd.Flags().NFlag() == 0 {
			//	return fmt.Errorf("'service measure' requires flag(s)")
			//	}
			return nil

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ksvcName := args[0]
			ksvcNSName := "default"
			if cmd.Flags().Changed("namespace") {
				ksvcNSName = n
			}
			f := NewFetcher(ksvcName, ksvcNSName, p)
			if err := f.GetKSVCResources(); err != nil {
				fmt.Printf("Failed: %v\n ", err)
			}
			defer dumpToTables(f, verbose)
			return nil
		},
	}

	serviceCmd.Flags().StringVarP(&n, "namespace", "n", "", "the target namespace")
	serviceCmd.Flags().StringVarP(&verbose, "verbose", "", "", "enable verbose output")
	return serviceCmd
}

func dumpToTables(f *Fetcher, verbose string) error {

	if f.KSVC == nil {
		return fmt.Errorf("Can't fetch any KSVC CR for the target service")
	}

	var table Table
	var err error

	switch strings.ToLower(verbose) {
	case "keyinfo":
		table = NewTable(os.Stdout, []string{"Resource Type", "Name", "Status", "KeyInfo"})
		//table.SetSeperator(false)
		err = createKeyInfoTable(table, f)
	case "conditions":
		table = NewTable(os.Stdout, []string{"Resource Type", "Name", "Status.Conditions"})
		table.SetSeperator(false)
		err = createConditionsTable(table, f)
	default:
		table = NewTable(os.Stdout, []string{"Resource Type", "Name", "Status", "Created At", "LastTransistion At"})
		table.SetSeperator(false)
		err = createTinyTable(table, f)
	}

	defer table.Print()

	return err
}

func createKeyInfoTable(table Table, f *Fetcher) error {

	ksvcResource := NewPrintableResource(0, "ksvc", f.KSVC.ObjectMeta.Name, strconv.FormatBool(f.KSVC.IsReady()), WithVerboseType(verbose))
	for key, value := range f.KSVC.Spec.Template.ObjectMeta.GetAnnotations() {
		ksvcResource.AppendKeyInfo([]string{key, value})
	}
	if f.KSVC.Spec.Template.Spec.ContainerConcurrency != nil {
		ksvcResource.AppendKeyInfo([]string{"containerConcurrency", strconv.FormatInt(*(f.KSVC.Spec.Template.Spec.ContainerConcurrency), 10)})
	}
	if f.KSVC.Spec.Template.Spec.EnableServiceLinks != nil {
		ksvcResource.AppendKeyInfo([]string{"enableServiceLinks", strconv.FormatBool(*(f.KSVC.Spec.Template.Spec.EnableServiceLinks))})
	}
	if f.KSVC.Spec.Template.Spec.TimeoutSeconds != nil {
		ksvcResource.AppendKeyInfo([]string{"timeoutSeconds", strconv.FormatInt(*(f.KSVC.Spec.Template.Spec.TimeoutSeconds), 10)})
	}
	table.AddMuitpleRows(ksvcResource.DumpResource())

	if f.Configuration == nil {
		return fmt.Errorf("Can't fetch any Configuration CR for the target service")
	}
	configResource := NewPrintableResource(1, "configuration", f.Configuration.ObjectMeta.Name, strconv.FormatBool(f.Configuration.IsReady()), WithVerboseType(verbose))
	if f.Configuration.Status.LatestCreatedRevisionName != "" {
		configResource.AppendKeyInfo([]string{"latestCreatedRevisionName", f.Configuration.Status.LatestCreatedRevisionName})
	}
	if f.Configuration.Status.LatestReadyRevisionName != "" {
		configResource.AppendKeyInfo([]string{"latestCreatedRevisionName", f.Configuration.Status.LatestReadyRevisionName})
	}
	table.AddMuitpleRows(configResource.DumpResource())

	if f.Revision == nil {
		return fmt.Errorf("Can't fetch any Revision CR for the target service")
	}
	revisionResource := NewPrintableResource(2, "revision", f.Revision.ObjectMeta.Name, strconv.FormatBool(f.Revision.IsReady()), WithVerboseType(verbose))
	table.AddMuitpleRows(revisionResource.DumpResource())

	if f.Deployment == nil {
		return fmt.Errorf("Can't fetch any Deployment for the target service")
	}
	deploymentResource := NewPrintableResource(3, "deployment", f.Deployment.ObjectMeta.Name, "--",
		WithVerboseType(verbose))
	for _, container := range f.Deployment.Spec.Template.Spec.Containers {
		deploymentResource.AppendKeyInfo([]string{container.Name + ":image-pull-policy", string(container.ImagePullPolicy)})

		if container.Resources.Limits.Cpu() != nil {
			deploymentResource.AppendKeyInfo([]string{container.Name + ":cpu-limit", container.Resources.Limits.Cpu().String()})
		}
		if container.Resources.Requests.Cpu() != nil {
			deploymentResource.AppendKeyInfo([]string{container.Name + ":cpu-request", string(container.Resources.Requests.Cpu().String())})
		}
		if container.Resources.Limits.Memory() != nil {
			deploymentResource.AppendKeyInfo([]string{container.Name + ":memory-limit", string(container.Resources.Limits.Memory().String())})
		}
		if container.Resources.Requests.Memory() != nil {
			deploymentResource.AppendKeyInfo([]string{container.Name + ":memory-request", string(container.Resources.Requests.Memory().String())})
		}
	}
	table.AddMuitpleRows(deploymentResource.DumpResource())

	if f.Images == nil {
		return fmt.Errorf("Can't fetch any ImageCache CR for the target service")
	}
	imgCacheResource := NewPrintableResource(3, "image", f.Images.ObjectMeta.Name, "--",
		WithVerboseType(verbose))
	table.AddMuitpleRows(imgCacheResource.DumpResource())

	if f.KPA == nil {
		return fmt.Errorf("Can't fetch any ImageCach CR for the target service")
	}
	kpaResource := NewPrintableResource(3, "podautoscaler", f.KPA.ObjectMeta.Name, strconv.FormatBool(f.KPA.IsReady()),
		WithVerboseType(verbose))
	kpaResource.AppendKeyInfo([]string{"actualScale", strconv.FormatInt((int64)(*(f.KPA.Status.ActualScale)), 10)})
	kpaResource.AppendKeyInfo([]string{"desiredScale", strconv.FormatInt((int64)(*(f.KPA.Status.DesiredScale)), 10)})
	for key, annotation := range f.KPA.Status.Annotations {
		kpaResource.AppendKeyInfo([]string{key, annotation})
	}

	table.AddMuitpleRows(kpaResource.DumpResource())

	if f.Metrics == nil {
		return fmt.Errorf("Can't fetch any ImageCach CR for the target service")
	}
	metricResource := NewPrintableResource(4, "metrics", f.Metrics.ObjectMeta.Name, strconv.FormatBool(f.Metrics.IsReady()),
		WithVerboseType(verbose))
	metricResource.AppendKeyInfo([]string{"scrapeTarget", string(f.Metrics.Spec.ScrapeTarget)})
	metricResource.AppendKeyInfo([]string{"panicWindow", f.Metrics.Spec.PanicWindow.String()})
	metricResource.AppendKeyInfo([]string{"stableWindow", f.Metrics.Spec.StableWindow.String()})
	table.AddMuitpleRows(metricResource.DumpResource())

	if f.SKS == nil {
		return fmt.Errorf("Can't fetch any SKS CR for the target service")
	}
	sksResource := NewPrintableResource(4, "sks", f.SKS.ObjectMeta.Name, fmt.Sprintf("%v", f.SKS.Spec.Mode),
		WithVerboseType(verbose))
	sksResource.AppendKeyInfo([]string{"numActivators", strconv.FormatInt((int64)(f.SKS.Spec.NumActivators), 10)})
	table.AddMuitpleRows(sksResource.DumpResource())

	publicsvcResource := NewPrintableResource(5, "svc", f.SKSPublicSVC.ObjectMeta.Name, "--",
		WithVerboseType(verbose))
	table.AddMuitpleRows(publicsvcResource.DumpResource())

	privatesvcResource := NewPrintableResource(5, "svc", f.SKSPrivateSVC.ObjectMeta.Name, "--",
		WithVerboseType(verbose))
	table.AddMuitpleRows(privatesvcResource.DumpResource())

	privateEPResource := NewPrintableResource(6, "endpoints", f.PrivateEendpoint.ObjectMeta.Name, "--",
		WithVerboseType(verbose))
	table.AddMuitpleRows(privateEPResource.DumpResource())

	publicEpResource := NewPrintableResource(5, "endpoints", f.SKSPublicEndpoint.ObjectMeta.Name, "--",
		WithVerboseType(verbose))
	table.AddMuitpleRows(publicEpResource.DumpResource())

	if f.Routes == nil {
		return fmt.Errorf("Can't fetch any Routes CR for the target service")
	}
	routesResource := NewPrintableResource(1, "routes", f.Routes.ObjectMeta.Name, strconv.FormatBool(f.Routes.IsReady()),
		WithVerboseType(verbose))
	for _, traffic := range f.Routes.Spec.Traffic {
		routesResource.AppendKeyInfo([]string{"traffic:configurationName", traffic.ConfigurationName})
		routesResource.AppendKeyInfo([]string{"traffic:latestRevision", strconv.FormatBool(*(traffic.LatestRevision))})
		routesResource.AppendKeyInfo([]string{"traffic:percent", strconv.FormatInt(*(traffic.Percent), 10)})
	}
	table.AddMuitpleRows(routesResource.DumpResource())

	if f.RoutesSVC == nil {
		return fmt.Errorf("Can't fetch any external public service  CR for the target service")
	}
	externalsvcResource := NewPrintableResource(2, "svc", f.RoutesSVC.ObjectMeta.Name, "--",
		WithVerboseType(verbose))
	table.AddMuitpleRows(externalsvcResource.DumpResource())

	if f.King == nil {
		return fmt.Errorf("Can't fetch any Ingress CR for the target service")
	}
	ingressResource := NewPrintableResource(2, "ingress", f.King.ObjectMeta.Name, strconv.FormatBool(f.King.IsReady()),
		WithVerboseType(verbose))
	table.AddMuitpleRows(ingressResource.DumpResource())

	return nil

}

func createConditionsTable(table Table, f *Fetcher) error {

	ksvcResource := NewPrintableResource(0, "ksvc", f.KSVC.ObjectMeta.Name, strconv.FormatBool(f.KSVC.IsReady()),
		WithCreatedAt(f.KSVC.CreationTimestamp.Rfc3339Copy().String()),
		WithVerboseType(verbose))
	for _, cond := range f.KSVC.Status.Conditions {
		ksvcResource.AppendConditions([]string{cond.LastTransitionTime.Inner.Rfc3339Copy().String(), string(cond.Type), string(cond.Status), cond.Reason, cond.Message})
	}
	table.AddMuitpleRows(ksvcResource.DumpResource())

	if f.Configuration == nil {
		return fmt.Errorf("Can't fetch any Configuration CR for the target service")
	}
	configResource := NewPrintableResource(1, "configuration", f.Configuration.ObjectMeta.Name, strconv.FormatBool(f.Configuration.IsReady()),
		WithCreatedAt(f.Configuration.CreationTimestamp.Rfc3339Copy().String()),
		WithVerboseType(verbose))
	for _, cond := range f.Configuration.Status.Conditions {
		configResource.AppendConditions([]string{cond.LastTransitionTime.Inner.Rfc3339Copy().String(), string(cond.Type), string(cond.Status), cond.Reason, cond.Message})
	}
	table.AddMuitpleRows(configResource.DumpResource())

	if f.Revision == nil {
		return fmt.Errorf("Can't fetch any Revision CR for the target service")
	}
	revisionResource := NewPrintableResource(2, "revision", f.Revision.ObjectMeta.Name, strconv.FormatBool(f.Revision.IsReady()),
		WithCreatedAt(f.Revision.CreationTimestamp.Rfc3339Copy().String()),
		WithVerboseType(verbose))
	for _, cond := range f.Revision.Status.Conditions {
		revisionResource.AppendConditions([]string{cond.LastTransitionTime.Inner.Rfc3339Copy().String(), string(cond.Type), string(cond.Status), cond.Reason, cond.Message})
	}
	table.AddMuitpleRows(revisionResource.DumpResource())

	if f.Deployment == nil {
		return fmt.Errorf("Can't fetch any Deployment for the target service")
	}
	deploymentResource := NewPrintableResource(3, "deployment", f.Deployment.ObjectMeta.Name, "--",
		WithCreatedAt(f.Deployment.CreationTimestamp.Rfc3339Copy().String()),
		WithVerboseType(verbose))
	for _, cond := range f.Deployment.Status.Conditions {
		deploymentResource.AppendConditions([]string{cond.LastTransitionTime.Rfc3339Copy().String(), string(cond.Type), string(cond.Status), cond.Reason, cond.Message})
	}
	table.AddMuitpleRows(deploymentResource.DumpResource())

	if f.Images == nil {
		return fmt.Errorf("Can't fetch any ImageCache CR for the target service")
	}
	imgCacheResource := NewPrintableResource(3, "image", f.Images.ObjectMeta.Name, "--",
		WithCreatedAt(f.Images.CreationTimestamp.Rfc3339Copy().String()),
		WithVerboseType(verbose))
	table.AddMuitpleRows(imgCacheResource.DumpResource())

	if f.KPA == nil {
		return fmt.Errorf("Can't fetch any ImageCach CR for the target service")
	}
	kpaResource := NewPrintableResource(3, "podautoscaler", f.KPA.ObjectMeta.Name, strconv.FormatBool(f.KPA.IsReady()),
		WithCreatedAt(f.KPA.CreationTimestamp.Rfc3339Copy().String()),
		WithVerboseType(verbose))
	for _, cond := range f.KPA.Status.Conditions {
		kpaResource.AppendConditions([]string{cond.LastTransitionTime.Inner.Rfc3339Copy().String(), string(cond.Type), string(cond.Status), cond.Reason, cond.Message})
	}
	table.AddMuitpleRows(kpaResource.DumpResource())

	if f.Metrics == nil {
		return fmt.Errorf("Can't fetch any ImageCach CR for the target service")
	}
	metricResource := NewPrintableResource(4, "metrics", f.Metrics.ObjectMeta.Name, strconv.FormatBool(f.Metrics.IsReady()),
		WithCreatedAt(f.Metrics.CreationTimestamp.Rfc3339Copy().String()),
		WithVerboseType(verbose))
	for _, cond := range f.Metrics.Status.Conditions {
		metricResource.AppendConditions([]string{cond.LastTransitionTime.Inner.Rfc3339Copy().String(), string(cond.Type), string(cond.Status), cond.Reason, cond.Message})
	}
	table.AddMuitpleRows(metricResource.DumpResource())

	if f.SKS == nil {
		return fmt.Errorf("Can't fetch any SKS CR for the target service")
	}
	sksResource := NewPrintableResource(4, "sks", f.SKS.ObjectMeta.Name, fmt.Sprintf("%v", f.SKS.Spec.Mode),
		WithCreatedAt(f.SKS.CreationTimestamp.Rfc3339Copy().String()),
		WithVerboseType(verbose))
	for _, cond := range f.SKS.Status.Conditions {
		metricResource.AppendConditions([]string{cond.LastTransitionTime.Inner.Rfc3339Copy().String(), string(cond.Type), string(cond.Status), cond.Reason, cond.Message})
	}
	table.AddMuitpleRows(sksResource.DumpResource())

	publicsvcResource := NewPrintableResource(5, "svc", f.SKSPublicSVC.ObjectMeta.Name, "--",
		WithCreatedAt(f.SKSPublicSVC.CreationTimestamp.Rfc3339Copy().String()),
		WithVerboseType(verbose))
	table.AddMuitpleRows(publicsvcResource.DumpResource())

	privatesvcResource := NewPrintableResource(5, "svc", f.SKSPrivateSVC.ObjectMeta.Name, "--",
		WithCreatedAt(f.SKSPrivateSVC.CreationTimestamp.Rfc3339Copy().String()),
		WithVerboseType(verbose))
	table.AddMuitpleRows(privatesvcResource.DumpResource())

	privateEPResource := NewPrintableResource(6, "endpoints", f.PrivateEendpoint.ObjectMeta.Name, "--",
		WithCreatedAt(f.PrivateEendpoint.CreationTimestamp.Rfc3339Copy().String()),
		WithVerboseType(verbose))
	table.AddMuitpleRows(privateEPResource.DumpResource())

	publicEpResource := NewPrintableResource(5, "endpoints", f.SKSPublicEndpoint.ObjectMeta.Name, "--",
		WithCreatedAt(f.SKSPublicEndpoint.CreationTimestamp.Rfc3339Copy().String()),
		WithVerboseType(verbose))
	table.AddMuitpleRows(publicEpResource.DumpResource())

	if f.Routes == nil {
		return fmt.Errorf("Can't fetch any Routes CR for the target service")
	}
	routesResource := NewPrintableResource(1, "routes", f.Routes.ObjectMeta.Name, strconv.FormatBool(f.Routes.IsReady()),
		WithCreatedAt(f.Routes.CreationTimestamp.Rfc3339Copy().String()),
		WithVerboseType(verbose))
	for _, cond := range f.Routes.Status.Conditions {
		routesResource.AppendConditions([]string{cond.LastTransitionTime.Inner.Rfc3339Copy().String(), string(cond.Type), string(cond.Status), cond.Reason, cond.Message})
	}
	table.AddMuitpleRows(routesResource.DumpResource())

	if f.RoutesSVC == nil {
		return fmt.Errorf("Can't fetch any external public service  CR for the target service")
	}
	externalsvcResource := NewPrintableResource(2, "svc", f.RoutesSVC.ObjectMeta.Name, "--",
		WithCreatedAt(f.RoutesSVC.CreationTimestamp.Rfc3339Copy().String()),
		WithVerboseType(verbose))
	table.AddMuitpleRows(externalsvcResource.DumpResource())

	if f.King == nil {
		return fmt.Errorf("Can't fetch any Ingress CR for the target service")
	}
	ingressResource := NewPrintableResource(2, "ingress", f.King.ObjectMeta.Name, strconv.FormatBool(f.King.IsReady()),
		WithCreatedAt(f.King.CreationTimestamp.Rfc3339Copy().String()),
		WithVerboseType(verbose))
	for _, cond := range f.King.Status.Conditions {
		ingressResource.AppendConditions([]string{cond.LastTransitionTime.Inner.Rfc3339Copy().String(), string(cond.Type), string(cond.Status), cond.Reason, cond.Message})
	}
	table.AddMuitpleRows(ingressResource.DumpResource())

	return nil

}

func createTinyTable(table Table, f *Fetcher) error {

	ksvcResource := NewPrintableResource(0, "ksvc", f.KSVC.ObjectMeta.Name, strconv.FormatBool(f.KSVC.IsReady()),
		WithCreatedAt(f.KSVC.CreationTimestamp.Rfc3339Copy().String()),
		WithLastTransitionAt(f.KSVC.Status.GetCondition(servingv1api.ServiceConditionReady).LastTransitionTime.Inner.Rfc3339Copy().String()))
	table.AddMuitpleRows(ksvcResource.DumpResource())

	if f.Configuration == nil {
		return fmt.Errorf("Can't fetch any Configuration CR for the target service")
	}
	configResource := NewPrintableResource(1, "configuration", f.Configuration.ObjectMeta.Name, strconv.FormatBool(f.Configuration.IsReady()),
		WithCreatedAt(f.Configuration.CreationTimestamp.Rfc3339Copy().String()),
		WithLastTransitionAt(f.Configuration.Status.GetCondition(servingv1api.ConfigurationConditionReady).LastTransitionTime.Inner.Rfc3339Copy().String()))
	table.AddMuitpleRows(configResource.DumpResource())

	if f.Revision == nil {
		return fmt.Errorf("Can't fetch any Revision CR for the target service")
	}
	revisionResource := NewPrintableResource(2, "revision", f.Revision.ObjectMeta.Name, strconv.FormatBool(f.Revision.IsReady()),
		WithCreatedAt(f.Revision.CreationTimestamp.Rfc3339Copy().String()),
		WithLastTransitionAt(f.Revision.Status.GetCondition(servingv1api.RevisionConditionReady).LastTransitionTime.Inner.Rfc3339Copy().String()))
	table.AddMuitpleRows(revisionResource.DumpResource())

	if f.Deployment == nil {
		return fmt.Errorf("Can't fetch any Deployment for the target service")
	}
	deploymentResource := NewPrintableResource(3, "deployment", f.Deployment.ObjectMeta.Name, "--",
		WithCreatedAt(f.Deployment.CreationTimestamp.Rfc3339Copy().String()))
	table.AddMuitpleRows(deploymentResource.DumpResource())

	if f.Images == nil {
		return fmt.Errorf("Can't fetch any ImageCache CR for the target service")
	}
	imgCacheResource := NewPrintableResource(3, "image", f.Images.ObjectMeta.Name, "--",
		WithCreatedAt(f.Images.CreationTimestamp.Rfc3339Copy().String()))
	table.AddMuitpleRows(imgCacheResource.DumpResource())

	if f.KPA == nil {
		return fmt.Errorf("Can't fetch any ImageCach CR for the target service")
	}
	kpaResource := NewPrintableResource(3, "podautoscaler", f.KPA.ObjectMeta.Name, strconv.FormatBool(f.KPA.IsReady()),
		WithCreatedAt(f.KPA.CreationTimestamp.Rfc3339Copy().String()),
		WithLastTransitionAt(f.KPA.Status.GetCondition(autoscalingv1v1alpha1api.PodAutoscalerConditionReady).LastTransitionTime.Inner.Rfc3339Copy().String()))
	table.AddMuitpleRows(kpaResource.DumpResource())

	if f.Metrics == nil {
		return fmt.Errorf("Can't fetch any ImageCach CR for the target service")
	}
	metricResource := NewPrintableResource(4, "metrics", f.Metrics.ObjectMeta.Name, strconv.FormatBool(f.Metrics.IsReady()),
		WithCreatedAt(f.Metrics.CreationTimestamp.Rfc3339Copy().String()),
		WithLastTransitionAt(f.Metrics.Status.GetCondition(autoscalingv1v1alpha1api.MetricConditionReady).LastTransitionTime.Inner.Rfc3339Copy().String()))
	table.AddMuitpleRows(metricResource.DumpResource())

	if f.SKS == nil {
		return fmt.Errorf("Can't fetch any SKS CR for the target service")
	}
	sksResource := NewPrintableResource(4, "sks", f.SKS.ObjectMeta.Name, fmt.Sprintf("%v", f.SKS.Spec.Mode),
		WithCreatedAt(f.SKS.CreationTimestamp.Rfc3339Copy().String()),
		WithLastTransitionAt(f.SKS.Status.GetCondition(nv1alpha1.ServerlessServiceConditionReady).LastTransitionTime.Inner.Rfc3339Copy().String()))
	table.AddMuitpleRows(sksResource.DumpResource())

	publicsvcResource := NewPrintableResource(5, "svc", f.SKSPublicSVC.ObjectMeta.Name, "--",
		WithCreatedAt(f.SKSPublicSVC.CreationTimestamp.Rfc3339Copy().String()))
	table.AddMuitpleRows(publicsvcResource.DumpResource())

	privatesvcResource := NewPrintableResource(5, "svc", f.SKSPrivateSVC.ObjectMeta.Name, "--",
		WithCreatedAt(f.SKSPrivateSVC.CreationTimestamp.Rfc3339Copy().String()))
	table.AddMuitpleRows(privatesvcResource.DumpResource())

	privateEPResource := NewPrintableResource(6, "endpoints", f.PrivateEendpoint.ObjectMeta.Name, "--",
		WithCreatedAt(f.PrivateEendpoint.CreationTimestamp.Rfc3339Copy().String()))
	table.AddMuitpleRows(privateEPResource.DumpResource())

	publicEpResource := NewPrintableResource(5, "endpoints", f.SKSPublicEndpoint.ObjectMeta.Name, "--",
		WithCreatedAt(f.SKSPublicEndpoint.CreationTimestamp.Rfc3339Copy().String()))
	table.AddMuitpleRows(publicEpResource.DumpResource())

	if f.Routes == nil {
		return fmt.Errorf("Can't fetch any Routes CR for the target service")
	}
	routesResource := NewPrintableResource(1, "routes", f.Routes.ObjectMeta.Name, strconv.FormatBool(f.Routes.IsReady()),
		WithCreatedAt(f.Routes.CreationTimestamp.Rfc3339Copy().String()),
		WithLastTransitionAt(f.Routes.Status.GetCondition(servingv1api.RouteConditionReady).LastTransitionTime.Inner.Rfc3339Copy().String()))
	table.AddMuitpleRows(routesResource.DumpResource())

	if f.RoutesSVC == nil {
		return fmt.Errorf("Can't fetch any external public service  CR for the target service")
	}
	externalsvcResource := NewPrintableResource(2, "svc", f.RoutesSVC.ObjectMeta.Name, "--",
		WithCreatedAt(f.RoutesSVC.CreationTimestamp.Rfc3339Copy().String()))
	table.AddMuitpleRows(externalsvcResource.DumpResource())

	if f.King == nil {
		return fmt.Errorf("Can't fetch any Ingress CR for the target service")
	}
	ingressResource := NewPrintableResource(2, "ingress", f.King.ObjectMeta.Name, strconv.FormatBool(f.King.IsReady()),
		WithCreatedAt(f.King.CreationTimestamp.Rfc3339Copy().String()),
		WithLastTransitionAt(f.King.Status.GetCondition(nv1alpha1.IngressConditionReady).LastTransitionTime.Inner.Rfc3339Copy().String()))
	table.AddMuitpleRows(ingressResource.DumpResource())

	return nil
}
