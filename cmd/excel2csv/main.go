package main

import (
	"flag"
	"fmt"
	"encoding/csv"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"log"
	"os"
)

const (
	NAME_COL_INDEX   = 0
	TARGET_ROW_INDEX = 0
)

func check(title string, err error) {
	if err != nil {
		e := fmt.Errorf("%s: %v", title, err)
		panic(e)
	}
}

func readSheet(excelFilePath, sheetName string) ([][]string, error) {
	excel, err := xlsx.OpenFile(excelFilePath)
	if err != nil {
		return nil, err
	}

	sheet := excel.Sheet[sheetName]
	if sheet == nil {
		return nil, fmt.Errorf("not found sheet: %s", sheetName)
	}

    records := make([][]string, 0, len(sheet.Rows))
    for r, _ := range sheet.Rows {
        cols := make([]string, 0, len(sheet.Cols))
        for c, _ := range sheet.Cols {
            cell := sheet.Cell(r, c)
            cols = append(cols, cell.Value)
        }
        records = append(records, cols)
    }
    return records, nil
}

func writeRecords(records [][]string, csvFile *os.File) error {
    return csv.NewWriter(csvFile).WriteAll(records)
}

func main() {
	var (
		excelFilePath = flag.String("excel", "variable.xlsx", "the excel file")
		sheetName     = flag.String("sheet", "Sheet1", "the sheet name in the excel")
        csvFilePath   = flag.String("csv", "-", "the csv file")
		verbose       = flag.Bool("v", false, "verbose")
	)
	flag.Usage = func() {
		o := flag.CommandLine.Output()
		cmd := os.Args[0]
		fmt.Fprintf(o, "Usage of %s:\n", cmd)
		fmt.Fprintf(o, "  %s converts the excel file into the csv file\n", cmd)
		flag.PrintDefaults()
		fmt.Fprintf(o, "example:\n")
		fmt.Fprintf(o, "  %s -excel input.xlsx -csv output.csv\n", cmd)
	}
	flag.Parse()

	if *verbose {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	records, err := readSheet(*excelFilePath, *sheetName)
	check("read excel", err)

    csvFile := os.Stdout
    if *csvFilePath != "-" {
        csvFile, err = os.Open(*csvFilePath)
        check("create csv", err)
    }

    err = writeRecords(records, csvFile)
    check("excel to csv", err)
}
