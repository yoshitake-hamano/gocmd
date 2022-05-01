package main

import (
	"flag"
	"fmt"
	"github.com/tealeg/xlsx"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"
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

func validateDirectory(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("is not directory: %s", path)
	}
	return nil
}

func createValiableMap(excelFilePath, sheetName, target string) (map[string]string, error) {
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

func createTemplatedFile(tpath, opath string, vm map[string]string) error {
	text, err := os.ReadFile(tpath)
	if err != nil {
		return err
	}

	tpl, err := template.New(opath).Parse(string(text))
	if err != nil {
		return err
	}

	out, err := os.Create(opath)
	defer out.Close()
	if err != nil {
		return err
	}
	if err := tpl.Execute(out, vm); err != nil {
		return err
	}
	return nil
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

	err := validateDirectory(*templateDirectory)
	check("template directory", err)

	err = validateDirectory(*outputDirectory)
	check("output directory", err)

	vm, err := createValiableMap(*excelFilePath, *sheetName, *target)
	check("valiable map", err)
	log.Printf("variable map = %v", vm)

	err = filepath.Walk(*templateDirectory, func(path string, info fs.FileInfo, err error) error {
		if path == *templateDirectory {
			return nil
		}
		relPath, err := filepath.Rel(*templateDirectory, path)
		if err != nil {
			return err
		}
		outputPath := filepath.Join(*outputDirectory, relPath)

		// is dir
		if info.IsDir() {
			err = validateDirectory(outputPath)
			if err == nil {
				// already have this directory(outputPath)
				return nil
			}
			err = os.Mkdir(outputPath, info.Mode())
			return err
		}

		// is file
		log.Printf("generating %s", outputPath)
		err = createTemplatedFile(path, outputPath, vm)
		check("create template files", err)

		return nil
	})
	check("walk template directory", err)
}
