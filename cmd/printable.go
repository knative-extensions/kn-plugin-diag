package main

import (
	"fmt"
	"os"
	"time"

	. "github.com/cdlliuy/knative-diagnose/pkg/utils"
)

type Resource struct {
	level            int
	typeName         string
	name             string
	ready            string
	createdAt        time.Time
	lastTransitionAt time.Time
	spec             []string
	status           []string
}

func NewResource(level int, typeName, name, ready string, createdAt, lastTransitionAt time.Time) *Resource {
	return &Resource{
		level:            level,
		typeName:         typeName,
		name:             name,
		ready:            ready,
		createdAt:        createdAt,
		lastTransitionAt: lastTransitionAt,
	}
}

func CombineTableData(orginal, appender [][]string) [][]string {
	for _, row := range appender {
		orginal = append(orginal, row)
	}
	return orginal
}

func printtable() {

	spec := NewTable(os.Stdout, []string{"", ""})

	specDetail := [][]string{
		[]string{"containerConcurrency", "100"},
		[]string{"timeout", "300"},
		[]string{"panicWindow", "60"},
		[]string{"mode", "proxy"},
	}

	for _, row := range specDetail {
		spec.Add(row)
	}
	dumpspec := spec.PrintDump()
	// for _, row := range dumpspec {
	// 	fmt.Println(row)
	// }

	// fmt.Println("")

	status := NewTable(os.Stdout, []string{"", "", ""})

	statusDetails := [][]string{
		[]string{"containerready", "true", "timestamp"},
		[]string{"imagepull", "true", "timestamp"},
		[]string{"active", "true", "timestamp"},
		[]string{"desiredScale", "1"},
		[]string{"metricsServiceName", "helloying-yxbqv-1-private"},
		[]string{"serviceName", "helloying"},
	}

	for _, row := range statusDetails {
		status.Add(row)
	}
	dumpStatus := status.PrintDump()
	// for _, row := range dumpStatus {
	// 	fmt.Println(row)
	// }

	// fmt.Println("")

	ksvcResource := NewResource(0, "ksvc", "name", "status", time.Now(), time.Now())
	configurationResource := NewResource(1, "configuration", "name", "status", time.Now(), time.Now())
	revisionResource := NewResource(2, "revision", "name", "status", time.Now(), time.Now())
	imageResource := NewResource(3, "image", "name", "status", time.Now(), time.Now())
	deploymentResource := NewResource(3, "deployment", "name", "status", time.Now(), time.Now())
	podResource := NewResource(4, "pod", "name", "status", time.Now(), time.Now())
	kpaResource := NewResource(3, "kpa", "name", "status", time.Now(), time.Now())
	metricResource := NewResource(4, "metrics", "name", "status", time.Now(), time.Now())
	sksResource := NewResource(4, "sks", "name", "status", time.Now(), time.Now())
	svcResource := NewResource(5, "svc", "name", "status", time.Now(), time.Now())
	endpointResource := NewResource(6, "endpoint", "name", "status", time.Now(), time.Now())
	routeResource := NewResource(1, "route", "name", "status", time.Now(), time.Now())
	ingressResource := NewResource(2, "ingress", "name", "status", time.Now(), time.Now())
	virtualservicesResource := NewResource(3, "virtualservice", "name", "status", time.Now(), time.Now())

	configurationResource.spec = dumpspec
	configurationResource.status = dumpStatus

	sksResource.spec = dumpspec
	sksResource.status = dumpStatus

	ingressResource.spec = dumpspec
	ingressResource.status = dumpStatus

	table := NewTable(os.Stdout, []string{"Resource Type", "Name", "Status", "Spec", "Status"})
	//table := NewTable(os.Stdout, []string{"Resource Type", "Name", "Status", "Created Time", "Last Ready Transistion Time"})
	//table := NewTable(os.Stdout, []string{"", "", "", "", ""})

	detail := [][]string{}
	detail = CombineTableData(detail, DumpResource(ksvcResource))
	detail = CombineTableData(detail, DumpResource(configurationResource))
	detail = CombineTableData(detail, DumpResource(revisionResource))
	detail = CombineTableData(detail, DumpResource(imageResource))
	detail = CombineTableData(detail, DumpResource(deploymentResource))
	detail = CombineTableData(detail, DumpResource(podResource))
	detail = CombineTableData(detail, DumpResource(kpaResource))
	detail = CombineTableData(detail, DumpResource(metricResource))
	detail = CombineTableData(detail, DumpResource(sksResource))
	detail = CombineTableData(detail, DumpResource(svcResource))
	detail = CombineTableData(detail, DumpResource(endpointResource))
	detail = CombineTableData(detail, DumpResource(routeResource))
	detail = CombineTableData(detail, DumpResource(ingressResource))
	detail = CombineTableData(detail, DumpResource(virtualservicesResource))

	// detail := [][]string{
	// 	[]string{"ksvc", "name", "status", "created at", "ready at"},

	// 	[]string{strings.Repeat(blanking, 1) + starting + "revision", "name", "status", "created at", "ready at"},
	// 	[]string{strings.Repeat(blanking, 2) + starting + "images", "name", "status", "created at", "ready at"},
	// 	[]string{strings.Repeat(blanking, 2) + starting + "deployment", "name", "status", "created at", "ready at"},
	// 	[]string{strings.Repeat(blanking, 3) + starting + "pod", "name", "status", "created at", "ready at"},
	// 	[]string{strings.Repeat(blanking, 4) + starting + "spec", "", "", "", "ready at"},
	// 	[]string{strings.Repeat(blanking, 2) + starting + "kpa", "name", "status", "created at", "ready at"},
	// 	[]string{strings.Repeat(blanking, 3) + starting + "metrics", "name", "status", "created at", "ready at"},
	// 	[]string{strings.Repeat(blanking, 3) + starting + "sks", "name", "status", "created at", "ready at"},
	// 	[]string{strings.Repeat(blanking, 4) + starting + "svc", "name", "status", "created at", "ready at"},
	// 	[]string{strings.Repeat(blanking, 5) + starting + "endpoint", "name", "status", "created at", "ready at"},
	// 	[]string{strings.Repeat(blanking, 0) + starting + "routes", "name", "status", "created at", "ready at"},
	// 	[]string{strings.Repeat(blanking, 1) + starting + "king", "name", "status", "created at", "ready at"},
	// 	[]string{strings.Repeat(blanking, 2) + starting + "virtualservices", "name", "status", "created at", "ready at"},
	// }

	for _, row := range detail {
		table.Add(row)
	}
	dumpTable := table.PrintDump()
	for _, row := range dumpTable {
		fmt.Println(row)
	}

}
