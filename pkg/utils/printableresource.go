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

package utils

import (
	"fmt"
	"os"
	"strings"

	. "knative.dev/kn-plugin-diag/pkg/models"

	"github.com/fatih/color"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type PrintableResource struct {
	level            int
	typeName         string
	name             string
	ready            string
	createdAt        string
	lastTransitionAt string
	keyInfo          [][]string
	conditions       [][]string
	verboseType      string
}

type Option func(PrintableResource) PrintableResource

func NewPrintableResource(level int, typeName, name string, options ...Option) *PrintableResource {
	res := PrintableResource{
		level:            level,
		typeName:         typeName,
		name:             name,
		ready:            "-",
		verboseType:      "",
		createdAt:        "",
		lastTransitionAt: "",
		keyInfo:          make([][]string, 0),
		conditions:       make([][]string, 0),
	}

	for _, option := range options {
		res = option(res)
	}
	return &res
}

func WithReady(ready string) Option {
	return func(res PrintableResource) PrintableResource {
		res.ready = ready
		return res
	}
}

func WithCreatedAt(createdAt string) Option {
	return func(res PrintableResource) PrintableResource {
		res.createdAt = createdAt
		return res
	}
}

func WithLastTransitionAt(lastTransitionAt string) Option {
	return func(res PrintableResource) PrintableResource {
		res.lastTransitionAt = lastTransitionAt
		return res
	}
}

func WithVerboseType(verboseType string) Option {
	return func(res PrintableResource) PrintableResource {
		res.verboseType = verboseType
		return res
	}
}

func (res *PrintableResource) appendKeyInfo(keyInfo []string) {
	res.keyInfo = append(res.keyInfo, keyInfo)
}

func (res *PrintableResource) appendConditions(conditions []string) {
	res.conditions = append(res.conditions, conditions)
}

func (res *PrintableResource) dumpSubTable(rows [][]string, requireSeperator bool) []string {

	col := 1
	for _, row := range rows {
		if len(row) > col {
			col = len(row)
		}
	}
	subtable := NewTable(os.Stdout, make([]string, col))
	subtable.SetSeperator(requireSeperator)
	subtable.SetHeaderPrinted(true)
	for _, row := range rows {
		subtable.Add(row)
	}
	return subtable.PrintDump(false)
}

func (res *PrintableResource) DumpResource() [][]string {

	startFirstLine := "|---"
	startSubLines := "    |"
	blanking := "    "

	paddingFirstLine := ""
	paddingSubLines := ""
	if res.level > 0 {
		paddingFirstLine = strings.Repeat(blanking, res.level-1) + startFirstLine
		paddingSubLines = strings.Repeat(blanking, res.level-1) + startSubLines
	} else {
		paddingFirstLine = ""
		paddingSubLines = "|"
	}

	data := [][]string{}
	switch res.verboseType {
	case "keyinfo":
		data = res.dumpToMultipleRows(res.keyInfo, paddingFirstLine, paddingSubLines, []string{res.typeName, res.name}...)
	default:
		data = res.dumpToMultipleRows(res.conditions, paddingFirstLine, paddingSubLines, []string{res.typeName, res.name, res.createdAt}...)
	}

	return data
}

func (res *PrintableResource) dumpToMultipleRows(elements [][]string, paddingFirstLine, paddingSubLines string, fixedColumns ...string) [][]string {

	data := [][]string{}
	var row []string
	subtableRows := res.dumpSubTable(elements, false)
	//empty conditions, we need to copy the fixed columns
	if len(subtableRows) == 0 {
		if len(fixedColumns) != 0 {
			row = []string{paddingFirstLine + fixedColumns[0]}
			for j := 1; j < len(fixedColumns); j++ {
				row = append(row, fixedColumns[j])
			}
		}
		row = append(row, "")
		data = append(data, row)
	} else {
		for i := 0; i < len(subtableRows); i++ {
			r := subtableRows[i]
			if i == 0 { //header line
				if len(fixedColumns) != 0 {
					row = []string{paddingFirstLine + fixedColumns[0]}
					for j := 1; j < len(fixedColumns); j++ {
						row = append(row, fixedColumns[j])
					}
				}
			} else {
				if len(fixedColumns) != 0 {
					row = []string{paddingSubLines}
					for j := 1; j < len(fixedColumns); j++ {
						row = append(row, "")
					}
				}
			}
			row = append(row, r) //append the subtable row info
			data = append(data, row)
		}
	}
	return data
}

func (res *PrintableResource) addKeyInfoRows(crName, key string, val interface{}) {

	cWarning := color.New(color.FgYellow).Add(color.Bold)

	switch vv := val.(type) {
	case []interface{}:
		//dump [].*
		cWarning.Printf("Detected slice for %s key %s, recommend to add [*] to retrieve nested fields\n", crName, key)
		for k, v := range vv {
			res.appendKeyInfo([]string{fmt.Sprintf("%s[%d]", key, k), fmt.Sprintf("%v", v)})
		}
	case map[string]interface{}:
		//dump map.*
		for nestedKey, nestedValue := range vv {
			res.appendKeyInfo([]string{fmt.Sprintf("%s.%s", key, nestedKey), fmt.Sprintf("%v", nestedValue)})
		}
	default:
		res.appendKeyInfo([]string{key, fmt.Sprintf("%v", val)})
	}

}

func (res *PrintableResource) addKeyInfoSliceDeepFirstRetrieve(keyPrefix string, object map[string]interface{}, depth int, segment []string, crName, objName string) error {

	if keyPrefix != "" {
		keyPrefix = keyPrefix + "." + segment[depth]
	} else {
		keyPrefix = segment[depth]
	}

	vv, ok, err := unstructured.NestedFieldNoCopy(object, strings.Split(segment[depth], ".")...)
	if !ok || err != nil {
		SayWarningMessage("Failed to load key info %s for %s %s, %v\n", keyPrefix, crName, objName, err)
		return nil
	}

	switch val := vv.(type) {
	case []interface{}:
		//handle xxx[*].yyy[*]...
		if depth == len(segment)-1 {
			for k, v := range val {
				res.addKeyInfoRows(crName, fmt.Sprintf("%s[%d]", keyPrefix, k), v)
			}
			return nil
		}
		for k, v := range val {
			switch vv := v.(type) {
			case map[string]interface{}:
				res.addKeyInfoSliceDeepFirstRetrieve(fmt.Sprintf("%s[%d]", keyPrefix, k), vv, depth+1, segment, crName, objName)
			default:
				fmt.Println("should go here222?")
			}
		}
	default:
		//handle xxx[*].yyy where yyy is a single value or map
		if depth == len(segment)-1 {
			res.addKeyInfoRows(crName, keyPrefix, val)
			return nil
		}
	}
	return nil
}

func (res *PrintableResource) AddKeyInfo(keyInfo []string, objectNode *ObjectNode) error {

	if objectNode == nil || objectNode.Object == nil || objectNode.Object.Object == nil {
		SayWarningMessage("Failed to load object for %s %s\n", objectNode.CRName, objectNode.ObjectName)
		return nil
	}

	object := objectNode.Object.Object
	for _, key := range keyInfo {
		if !strings.Contains(key, "[*]") {
			//no slice included
			val, ok, err := unstructured.NestedFieldNoCopy(object, strings.Split(key, ".")...)
			if !ok || err != nil {
				SayWarningMessage("Failed to load key info %s for %s %s, %v\n", key, objectNode.CRName, objectNode.ObjectName, err)
				continue
			}
			res.addKeyInfoRows(objectNode.CRName, key, val)
		} else {
			//handle slice output , defined in keyinfo by [*]
			segment := strings.Split(key, "[*]")
			for k, v := range segment {
				segment[k] = strings.TrimPrefix(v, ".")
			}
			if segment[len(segment)-1] == "" {
				segment = segment[0 : len(segment)-1]
			}
			err := res.addKeyInfoSliceDeepFirstRetrieve("", object, 0, segment, objectNode.CRName, objectNode.ObjectName)
			if err != nil {
				SayWarningMessage("Failed to load key info %s for %s %s, %v\n", key, objectNode.CRName, objectNode.ObjectName, err)
				continue
			}
		} //end of else
	}

	return nil

}

func (res *PrintableResource) AddConditions(objectNode *ObjectNode, conditionInfos map[string][]ConditionInfo) error {

	if objectNode == nil || objectNode.Object == nil || objectNode.Object.Object == nil {
		SayWarningMessage("Failed to load object for %s %s\n", objectNode.CRName, objectNode.ObjectName)
		return nil
	}

	object := objectNode.Object.Object
	conditions, ok, err := unstructured.NestedSlice(object, strings.Split("status.conditions", ".")...)
	if !ok || err != nil {
		if conditionInfo, ok := conditionInfos[objectNode.CRName]; ok && len(conditionInfo) != 0 {
			SayWarningMessage("Failed to load the status.conditions for %s %s, %v\n", objectNode.CRName, objectNode.ObjectName, err)
			return nil
		}
		return nil
	}

	//for fast retrieve
	conditionMaps := make(map[string]map[string]interface{})
	for _, condition := range conditions {
		if m, ok := condition.(map[string]interface{}); ok {
			if _, ok := m["type"]; ok {
				conditionMaps[fmt.Sprintf("%v", m["type"])] = m
			} else {
				SayWarningMessage("Invalid structure for %s status.conditions %v\n", objectNode.CRName, m)
				return nil
			}
		} else {
			SayWarningMessage("Invalid structure for %s status.conditions %v\n", objectNode.CRName, condition)
			return nil
		}
	}

	//if defined in conditionInfo map, then sort the condition output and check abnormal status from conditionInfo map
	//if len(conditionInfo) != len(conditions),  the deifnition of conditionInfo is not accurate.
	if conditionInfo, ok := conditionInfos[objectNode.CRName]; ok && len(conditionInfo) != 0 && len(conditionInfo) == len(conditionMaps) {
		for _, v := range conditionInfo {
			if m, ok := conditionMaps[v.Type]; ok {
				if _, ok := m["status"]; ok && len(v.Expected) != 0 && strings.Index(v.Expected, fmt.Sprintf("%v", m["status"])) == -1 {
					res.addConditionRows(m, false)
				} else {
					res.addConditionRows(m)
				}
			} else {
				//ignore the condition.type if not defined in conditionInfo or status.conditions
			}
		}
	} else {
		for _, condition := range conditions {
			res.addConditionRows(condition)
		}
	}

	return nil

}

func (res *PrintableResource) addConditionRows(condition interface{}, asExpected ...bool) {

	c := color.New(color.FgRed).Add(color.Bold)

	if m, ok := condition.(map[string]interface{}); ok {
		cons := []string{}
		if _, ok := m["lastTransitionTime"]; ok {
			cons = append(cons, fmt.Sprintf("%v", m["lastTransitionTime"]))
		}
		if _, ok := m["type"]; ok {
			if len(asExpected) != 0 && !asExpected[0] {
				cons = append(cons, c.Sprintf("%v", m["type"]))
			} else {
				cons = append(cons, fmt.Sprintf("%v", m["type"]))
			}
		}
		if _, ok := m["status"]; ok {
			if len(asExpected) != 0 && !asExpected[0] {
				cons = append(cons, c.Sprintf("%v", m["status"]))
			} else {
				cons = append(cons, fmt.Sprintf("%v", m["status"]))
			}
		}
		if _, ok := m["reason"]; ok {
			cons = append(cons, fmt.Sprintf("%v", m["reason"]))
		}
		if _, ok := m["message"]; ok {
			cons = append(cons, fmt.Sprintf("%v", m["message"]))
		}
		res.appendConditions(cons)
	}
}
