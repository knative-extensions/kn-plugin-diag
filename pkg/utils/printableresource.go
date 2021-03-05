package utils

import (
	"os"
	"strings"
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

func NewPrintableResource(level int, typeName, name, ready string, options ...Option) *PrintableResource {
	res := PrintableResource{
		level:            level,
		typeName:         typeName,
		name:             name,
		ready:            ready,
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

func (res *PrintableResource) AppendKeyInfo(keyInfo []string) {
	res.keyInfo = append(res.keyInfo, keyInfo)
}

func (res *PrintableResource) AppendConditions(conditions []string) {
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
	return subtable.PrintDump()
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
	var row []string
	switch res.verboseType {
	case "keyinfo":
		keyInfo := res.dumpSubTable(res.keyInfo, false)
		if len(keyInfo) == 0 {
			row = []string{paddingFirstLine + res.typeName, res.name, ""}
			data = append(data, row)
		} else {
			for i := 0; i < len(keyInfo); i++ {
				k := keyInfo[i]
				var row []string
				if i == 0 {
					row = []string{paddingFirstLine + res.typeName, res.name, res.ready, k}
				} else {
					row = []string{paddingSubLines, "", "", k}
				}
				data = append(data, row)
			}
		}
	case "conditions":
		conditions := res.dumpSubTable(res.conditions, false)
		if len(conditions) == 0 {
			row = []string{paddingFirstLine + res.typeName, res.name, ""}
			data = append(data, row)
		} else {
			for i := 0; i < len(conditions); i++ {
				c := conditions[i]
				if i == 0 {
					row = []string{paddingFirstLine + res.typeName, res.name, c}
				} else {
					row = []string{paddingSubLines, "", c}
				}
				data = append(data, row)
			}
		}
	default:
		row = []string{paddingFirstLine + res.typeName, res.name, res.ready, res.createdAt, res.lastTransitionAt}
		data = append(data, row)
	}

	return data
}
