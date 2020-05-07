package actions

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/jinzhu/gorm"
	"github.com/urfave/cli"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

func ImportReports(db *gorm.DB) func(c *cli.Context) {
	return func(c *cli.Context) {
		if c.NArg() < 0 {
			panic("Import Reports needs path")
		}
		dir := c.Args().Get(0)
		files := LoadFiles(dir)

		ExtractDataFromFiles(files)
	}
}

func LoadFiles(dir string) []*excelize.File {
	// TODO: Check if directory is empty
	var filepaths []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		filepaths = append(filepaths, path)
		if err != nil {
			fmt.Println(err)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	var files []*excelize.File
	sem := make(chan struct{}, 20)
	var wg sync.WaitGroup
	wg.Add(len(filepaths))
	for _, path := range filepaths {
		go func() {
			sem <- struct{}{}
			defer func() { <-sem }()
			defer wg.Done()

			f, err := excelize.OpenFile(path)
			if err != nil {
				fmt.Println(err)
			}
			files = append(files, f)
		}()
	}
	wg.Wait()

	return files
}

func ExtractDataFromFiles(files []*excelize.File) {
	sheetNames := map[string]int{
		"general information": 1,
		"financial position":  2,
		"profit or loss":      3,
	}

	sem := make(chan struct{}, 20)
	var wg sync.WaitGroup
	wg.Add(len(files))
	for _, f := range files {
		go func() {
			sem <- struct{}{}
			defer func() { <-sem }()
			defer wg.Done()

			// Names of the sheets in the file
			// Needed to access the sheets
			sheetMap := ArrangeWorksheets(f.GetSheetMap())

			// TODO: Get company name

			for key := range sheetNames {
				sheetIndex, ok := sheetNames[key]
				if !ok {
					fmt.Errorf("sheet name: %s does not exist", key)
				}

				row := 5
				titleCol := "D"
				amountCol := "B"
				for {
					excelSheetName := sheetMap[sheetIndex]

					titleCell := fmt.Sprintf("%s%d", titleCol, row)
					title := f.GetCellValue(excelSheetName, titleCell)
					if title == "" {
						break
					}

					amountCell := fmt.Sprintf("%s%d", amountCol, row)
					amount := f.GetCellValue(excelSheetName, amountCell)

					fmt.Printf("%s -- %s \n", title, amount)

					// TODO: Save Title And Amount to DB

					row++
				}
			}
		}()
	}
	wg.Wait()
}

// ArrangeWorksheets -- Arrange the worksheet names. sort in ascending order by index from excelize.File.GetSheetMap()
// originally the sheet has inconsistent numerical index e.g 1, 3, 6, 8...
// This function rearrange them using consistent index 0, 1, 2, 3...
func ArrangeWorksheets(worksheetMap map[int]string) []string {
	var worksheets []string
	var indexes []int
	for index := range worksheetMap {
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)
	for _, idx := range indexes {
		worksheets = append(worksheets, worksheetMap[idx])
	}
	return worksheets
}
