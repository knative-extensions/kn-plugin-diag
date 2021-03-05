package utils

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/fatih/color"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

type Table interface {
	Add(row []string)
	AddMuitpleRows(rows [][]string)
	Print()
	PrintDump() []string
	SetSeperator(bool)
	SetHeaderPrinted(bool)
}

type PrintableTable struct {
	writer           io.Writer
	headers          []string
	headerPrinted    bool
	maxSizes         []int
	rows             [][]string
	requireSeperator bool
	terminalWidth    int
}

func NewTable(w io.Writer, headers []string) Table {
	x, _ := terminal.Width()
	return &PrintableTable{
		writer:           w,
		headers:          headers,
		maxSizes:         make([]int, len(headers)),
		requireSeperator: true,
		terminalWidth:    int(x),
	}
}

func (t *PrintableTable) SetSeperator(required bool) {
	t.requireSeperator = required
}

func (t *PrintableTable) SetHeaderPrinted(headerPrinted bool) {
	t.headerPrinted = headerPrinted
}

func (t *PrintableTable) Add(row []string) {
	t.rows = append(t.rows, row)
}

func (t *PrintableTable) AddMuitpleRows(rows [][]string) {
	for _, row := range rows {
		t.rows = append(t.rows, row)
	}
	if t.requireSeperator {
		t.rows = append(t.rows, []string{})
	}
}

func (t *PrintableTable) PrintDump() []string {
	for _, row := range append(t.rows, t.headers) {
		t.calculateMaxSize(row)
	}

	if t.headerPrinted == false {
		t.printHeader()
		t.headerPrinted = true
	}

	dumps := []string{}
	maxSize := 0
	for _, row := range t.rows {
		if len(row) != 0 {
			line := t.printRowString(row)
			if len(line) > maxSize {
				maxSize = len(line)
			}
			if len(line) > t.terminalWidth {
				dumps = append(dumps, line[0:t.terminalWidth])
				indexOfSubline := strings.Index(line, "| ")
				indexOfHeaderline := strings.Index(line, "|---") + 4
				if indexOfSubline != -1 {
					dumps = append(dumps, line[0:indexOfSubline]+"|"+strings.Repeat(" ", 2*t.terminalWidth-len(line)-indexOfSubline-1)+line[t.terminalWidth:])
				} else {
					dumps = append(dumps, strings.Repeat(" ", indexOfHeaderline)+"|"+strings.Repeat(" ", 2*t.terminalWidth-len(line)-indexOfHeaderline-1)+line[t.terminalWidth:])
				}
			} else {
				dumps = append(dumps, line)
			}
		} else {
			dumps = append(dumps, "")
		}
	}

	if maxSize > t.terminalWidth {
		maxSize = t.terminalWidth
	}

	if t.requireSeperator {
		ret := []string{}
		ret = append(ret, strings.Repeat("-", maxSize)) //for header
		for _, row := range dumps {
			if len(row) != 0 {
				ret = append(ret, row)
			} else {
				ret = append(ret, strings.Repeat("-", maxSize))
			}
		}
		dumps = ret
	}

	t.rows = [][]string{}
	return dumps
}

func (t *PrintableTable) Print() {
	for _, row := range t.PrintDump() {
		fmt.Fprintln(t.writer, row)
	}
}

func (t *PrintableTable) calculateMaxSize(row []string) {
	for index, value := range row {
		cellLength := utf8.RuneCountInString(value)
		if t.maxSizes[index] < cellLength {
			t.maxSizes[index] = cellLength
		}
	}
}

func (t *PrintableTable) printHeader() {
	output := ""
	for col, value := range t.headers {
		output = output + t.cellValue(col, value)
	}
	output = strings.TrimRight(output, "| ")
	//skip colorized when output to file
	if t.writer == os.Stdout {
		c := color.New(color.FgWhite).Add(color.Bold)
		c.Fprintln(t.writer, output)
	} else {
		fmt.Fprintln(t.writer, output)
	}

}

func (t *PrintableTable) printRow(row []string) {
	output := ""
	for columnIndex, value := range row {
		if columnIndex == 0 {
			value = value
		}
		output = output + t.cellValue(columnIndex, value)
	}
	output = strings.TrimRight(output, "| ")
	fmt.Fprintln(t.writer, output)
}

func (t *PrintableTable) printRowString(row []string) string {
	output := ""
	for columnIndex, value := range row {
		if columnIndex == 0 {
			value = value
		}
		output = output + t.cellValue(columnIndex, value)
	}
	output = strings.TrimRight(output, "| ")
	return fmt.Sprintf(output)
}

func (t *PrintableTable) cellValue(col int, value string) string {
	padding := ""
	if col < len(t.headers)-1 {
		padding = strings.Repeat(" ", t.maxSizes[col]-utf8.RuneCountInString(value))
	}
	if t.requireSeperator {
		return fmt.Sprintf("%s%s    | ", value, padding)
	}
	return fmt.Sprintf("%s%s      ", value, padding)
}
