package main

import (
	"flag"
	"fmt"
	"github.com/tealeg/xlsx"
	"hash/crc32"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func check(title string, err error) {
	if err != nil {
		e := fmt.Errorf("%s: %v", title, err)
		panic(e)
	}
}

type HeadType string

const (
	Id            HeadType = "Id"
	Class         HeadType = "Class"
	Title         HeadType = "Title"
	IsSameRow     HeadType = "Is same row"
	DependOn      HeadType = "Depend on"
	Start         HeadType = "Start"
	End           HeadType = "End"
	CompletedRate HeadType = "Completed rate"
)

type ClassType string

const (
	Milestone ClassType = "Milestone"
	Separator ClassType = "Separator"
	Task      ClassType = "Task"
)

type Elem struct {
	id            string
	class         ClassType
	title         string
	isSameRow     bool
	dependOn      []string
	start         time.Time
	end           time.Time
	completedRate int
}

func plantUmlId(e Elem) string {
	return fmt.Sprintf("[N(%s,%s)]", e.title, e.id)
}

func createPlantUmlSeparator(e Elem, buf *strings.Builder) {
	buf.WriteString("-- ")
	buf.WriteString(e.title)
	buf.WriteString(" --\n")
}

func createPlantUmlMilestone(e Elem, buf *strings.Builder) {
	buf.WriteString(plantUmlId(e))
	buf.WriteString(" happens ")
	buf.WriteString(plantUmlTime(e.start))
	buf.WriteString("\n")
}

func createPlantUmlTask(e Elem, buf *strings.Builder) {
	buf.WriteString(plantUmlId(e))
	buf.WriteString(" starts ")
	buf.WriteString(plantUmlTime(e.start))
	buf.WriteString(" and ends ")
	buf.WriteString(plantUmlTime(e.end))
	buf.WriteString(" and is ")
	buf.WriteString(strconv.Itoa(e.completedRate))
	buf.WriteString("%% completed")
	buf.WriteString("\n")
}

func findEarliestStartTime(elems []Elem) (time.Time, error) {
	var t time.Time
	for _, e := range elems {
		if e.start.IsZero() {
			continue
		}
		if t.IsZero() {
			t = e.start
		}
		if t.After(e.start) {
			t = e.start
		}
	}
	return t, nil
}

func findId(elems []Elem, id string) (Elem, error) {
	for _, e := range elems {
		if e.id == id {
			return e, nil
		}
	}
	var dummy Elem
	return dummy, fmt.Errorf("not found id: %s", id)
}

func plantUmlTime(t time.Time) string {
	return t.Format("2006-01-02")
}

func createPlantUmlGantt(elems []Elem) (string, error) {
	var buf strings.Builder
	buf.WriteString(`
@startgantt
language ja
' printscale weekly
saturday are colored in LightGray
sunday   are colored in LightGray
' saturday are closed
' sunday   are closed

' Gantt aliases work differently from other diagrams (it is not possible to give the same "label" to a task)
' https://forum.plantuml.net/12176/gantt-aliases-differently-other-diagrams-possible-same-label
!define N(x,n) x<size:0>n</size>

`)
	projectStart, _ := findEarliestStartTime(elems)
	buf.WriteString("Project starts ")
	buf.WriteString(plantUmlTime(projectStart))
	buf.WriteString("\n\n")

	for _, e := range elems {
		if e.class == Milestone {
			createPlantUmlMilestone(e, &buf)
		}
		if e.class == Separator {
			createPlantUmlSeparator(e, &buf)
		}
		if e.class == Task {
			createPlantUmlTask(e, &buf)
		}
	}

	buf.WriteString("\n' displays on same row as\n")
	length := len(elems)
	for i, e := range elems {
		if (i + 1) == length {
			// skip last elem
			continue
		}
		if e.class != Task && e.class != Milestone {
			continue
		}
		if !elems[i+1].isSameRow {
			continue
		}
		buf.WriteString(plantUmlId(e))
		buf.WriteString(" displays on same row as ")
		buf.WriteString(plantUmlId(elems[i+1]))
		buf.WriteString("\n")
	}

	buf.WriteString("\n' relational ship\n")
	for _, e := range elems {
		if len(e.dependOn) == 0 {
			continue
		}
		for _, dep := range e.dependOn {
			depElem, err := findId(elems, dep)
			if err != nil {
				return buf.String(), fmt.Errorf("not found depend on id: %s", dep)
			}
			buf.WriteString(plantUmlId(depElem))
			buf.WriteString(" -> ")
			buf.WriteString(plantUmlId(e))
			buf.WriteString("\n")
		}
	}

	buf.WriteString("@endgantt")
	return buf.String(), nil
}

func getColIndex(headCells []*xlsx.Cell, h HeadType) (int, error) {
	for i, e := range headCells {
		if e.Value == string(h) {
			return i, nil
		}
	}
	return 0, fmt.Errorf("not found head: %s", h)
}

func getClassType(t string) (ClassType, error) {
	if string(Milestone) == t {
		return Milestone, nil
	}
	if string(Separator) == t {
		return Separator, nil
	}
	if string(Task) == t {
		return Task, nil
	}
	return Task, fmt.Errorf("unknown class type: %s", t)
}

func readExcel(excelFilePath, sheetName string) ([]Elem, error) {
	excel, err := xlsx.OpenFile(excelFilePath)
	if err != nil {
		return nil, err
	}

	sheet := excel.Sheet[sheetName]
	if sheet == nil {
		return nil, fmt.Errorf("not found sheet: %s", sheetName)
	}

	head := sheet.Rows[0]
	idIndex, err := getColIndex(head.Cells, Id)
	if err != nil {
		return nil, err
	}
	classIndex, err := getColIndex(head.Cells, Class)
	if err != nil {
		return nil, err
	}
	titleIndex, err := getColIndex(head.Cells, Title)
	if err != nil {
		return nil, err
	}
	isSameRowIndex, err := getColIndex(head.Cells, IsSameRow)
	if err != nil {
		return nil, err
	}
	dependOnIndex, err := getColIndex(head.Cells, DependOn)
	if err != nil {
		return nil, err
	}
	startIndex, err := getColIndex(head.Cells, Start)
	if err != nil {
		return nil, err
	}
	endIndex, err := getColIndex(head.Cells, End)
	if err != nil {
		return nil, err
	}
	completedRateIndex, err := getColIndex(head.Cells, CompletedRate)
	if err != nil {
		return nil, err
	}

	arr := make([]Elem, 0, sheet.MaxRow)
	for i, r := range sheet.Rows {
		if i == 0 {
			continue
		}
		id := ""
		if idIndex < len(r.Cells) {
			id = r.Cells[idIndex].Value
		}
		if id == "" {
			seed := fmt.Sprintf("%d%v", i, r)
			crc := crc32.ChecksumIEEE([]byte(seed))
			id = strconv.FormatUint(uint64(crc), 16)
		}
		if classIndex >= len(r.Cells) {
			continue
		}
		ct, err := getClassType(r.Cells[classIndex].Value)
		if err != nil {
			// skip unknown class
			continue
		}

		var dor []string
		if dependOnIndex < len(r.Cells) &&
			r.Cells[dependOnIndex].Value != "" {
			dor = strings.Split(r.Cells[dependOnIndex].Value, ",")
		}
		isr := 0
		if isSameRowIndex < len(r.Cells) {
			isr, _ = strconv.Atoi(r.Cells[isSameRowIndex].Value)
		}
		var sr time.Time
		if startIndex < len(r.Cells) {
			sr, _ = r.Cells[startIndex].GetTime(false)
		}
		var er time.Time
		if endIndex < len(r.Cells) {
			er, _ = r.Cells[endIndex].GetTime(false)
		}
		cr := 0
		if completedRateIndex < len(r.Cells) {
			cr, _ = strconv.Atoi(r.Cells[completedRateIndex].Value)
		}
		e := Elem{
			id:            id,
			class:         ct,
			title:         r.Cells[titleIndex].Value,
			isSameRow:     isr != 0,
			dependOn:      dor,
			start:         sr,
			end:           er,
			completedRate: cr,
		}
		arr = append(arr, e)
	}
	return arr, nil
}

func main() {
	var (
		excelFilePath = flag.String("excel", "sched.xlsx", "the excel file path which specify schedule")
		sheetName     = flag.String("sheet", "Sheet1", "the sheet name in the variable excel")
		verbose       = flag.Bool("v", false, "verbose")
		version       = flag.Bool("version", false, "version")
	)
	cmd := os.Args[0]
	flag.Usage = func() {
		o := flag.CommandLine.Output()
		fmt.Fprintf(o, "Usage of %s:\n", cmd)
		fmt.Fprintf(o, "  %s creates plantuml gantt source code from the schedule excel\n", cmd)
		flag.PrintDefaults()
		fmt.Fprintf(o, "example:\n")
		fmt.Fprintf(o, "  %s -excel sched.xlsx -sheet Sheet1\n", cmd)
	}
	flag.Parse()

	if *version {
		fmt.Printf("%s 0.3.0\n", cmd)
		return
	}

	if *verbose {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	excel, err := readExcel(*excelFilePath, *sheetName)
	check("readExcel()", err)
	gantt, err := createPlantUmlGantt(excel)
	check("createPlantUmlGantt()", err)

	fmt.Printf(gantt)
}
