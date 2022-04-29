package main

import (
	"flag"
	"fmt"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"log"
	"os"
)

const (
	NAME_COL_INDEX   = 0
	TARGET_ROW_INDEX = 0
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func valiableMap(excelFilePath, sheetName, target string) (map[string]string, error) {
	excel, err := xlsx.OpenFile(excelFilePath)
	if err != nil {
		return nil, err
	}

	sheet := excel.Sheet[sheetName]
	if sheet == nil {
		return nil, fmt.Errorf("not found sheet: %s", sheetName)
	}

	idxTarget := -1
	for i, _ := range sheet.Cols {
		cell := sheet.Cell(TARGET_ROW_INDEX, i)
		log.Printf("cell %s", cell.Value)
		if i == NAME_COL_INDEX {
			// skip name column
			continue
		}
		if cell.Value == target {
			idxTarget = i
			break
		}
	}
	if idxTarget == -1 {
		return nil, fmt.Errorf("not found target: %s", target)
	}

	vm := make(map[string]string)
	for i, _ := range sheet.Rows {
		if i == TARGET_ROW_INDEX {
			// skip target row
			continue
		}
		nameCell := sheet.Cell(i, NAME_COL_INDEX)
		valueCell := sheet.Cell(i, idxTarget)
		vm[nameCell.Value] = valueCell.Value
	}
	return vm, nil
}

func main() {
	var (
		excelFilePath     = flag.String("excel", "variable.xlsx", "the excel file path which defines template variable")
		sheetName         = flag.String("sheet", "Sheet1", "the sheet name in the variable excel")
		target            = flag.String("target", "", "the targer column in the variable excel")
		templateDirectory = flag.String("template", "", "the directory which has template files")
		outputDirectory   = flag.String("output", ".", "the output directory")
		verbose           = flag.Bool("v", false, "verbose")
	)
	flag.Usage = func() {
		o := flag.CommandLine.Output()
		cmd := os.Args[0]
		fmt.Fprintf(o, "Usage of %s:\n", cmd)
		fmt.Fprintf(o, "  %s creates files from the variable excel and the template files\n", cmd)
		flag.PrintDefaults()
		fmt.Fprintf(o, "example:\n")
		fmt.Fprintf(o, "  %s -excel variable.xlsx -target TARGET1 -template template -o build\n", cmd)
	}
	flag.Parse()

	if *verbose {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	vm, err := valiableMap(*excelFilePath, *sheetName, *target)
	check(err)
	log.Printf("excel    = %s", *excelFilePath)
	log.Printf("sheet    = %s", *sheetName)
	log.Printf("target   = %s", *target)
	log.Printf("template = %s", *templateDirectory)
	log.Printf("output   = %s", *outputDirectory)
	log.Printf("vm       = %v", vm)
}
