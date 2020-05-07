package actions

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/kevinjanada/magic-formula-cli/helpers"
	"github.com/kevinjanada/magic-formula-cli/models"
	"github.com/urfave/cli"
	"net/http"
	"path/filepath"
	"sync"
)

func DownloadReports(db *gorm.DB) func(c *cli.Context) {
	return func(c *cli.Context) {
		if c.String("year") == "" || c.String("period") == "" {
			panic("Download report needs year and period as arguments")
		}

		var stocks []*models.Stock
		if err := db.Find(&stocks).Error; err != nil {
			panic(err)
		}

		year := c.String("year")
		period := c.String("period")

		// Limit concurrency to 12
		sem := make(chan struct{}, 20)

		var wg sync.WaitGroup
		wg.Add(len(stocks))
		for _, stock := range stocks {
			go fetchAndDownloadReports(stock.Code, year, period, &wg, sem)
		}
		wg.Wait()
	}
}

func fetchAndDownloadReports(stockCode string, year string, period string, wg *sync.WaitGroup, sem chan struct{}) {
	sem <- struct{}{}
	defer func() { <-sem }()

	defer wg.Done()

	URL := generateURL(1, 1, year, period, stockCode)

	resp, err := http.Get(URL)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	finRep := &FinancialReportAPIResponse{Year: year, Period: fmt.Sprintf("trimester_%s", period)}
	err = helpers.JSONToStruct(resp, finRep)
	if err != nil {
		fmt.Println(err)
		return
	}
	finRep.DownloadExcelReports()
}

func generateURL(page int, pageSize int, year string, period string, emitenCode string) string {
	return fmt.Sprintf(
		"https://www.idx.co.id/umbraco/Surface/ListedCompany/GetFinancialReport?indexFrom=%d&pageSize=%d&year=%s&reportType=rdf&periode=tw%s&kodeEmiten=%s",
		page,
		pageSize,
		year,
		period,
		emitenCode,
	)
}

// FinancialReportAPIResponse -- Json response from IDX get financial report API
type FinancialReportAPIResponse struct {
	Year        string
	Period      string
	Search      Search   `json:"Search"`
	ResultCount int      `json:"ResultCount"`
	Results     []Result `json:"Results"`
}

// Search -- FRAPIResponseSearch field
type Search struct {
	ReportType string `json:"ReportType"`
	KodeEmiten string `json:"KodeEmiten"`
	Year       string `json:"Year"`
	Periode    string `json:"Periode"`
	Indexfrom  int    `json:"indexfrom"`
	Pagesize   int    `json:"pagesize"`
}

// Result -- FRAPIResponseSearch Results
type Result struct {
	KodeEmiten   string       `json:"KodeEmiten"`
	FileModified string       `json:"File_Modified"`
	ReportPeriod string       `json:"Report_Period"`
	ReportYear   string       `json:"Report_Year"`
	NamaEmiten   string       `json:"NamaEmiten"`
	Attachments  []Attachment `json:"Attachments"`
}

// Attachment -- FRAPIResponseAttachment object
type Attachment struct {
	EmitenCode   string `json:"Emiten_Code"`
	FileID       string `json:"File_ID"`
	FileModified string `json:"File_Modified"`
	FileName     string `json:"File_Name"`
	FilePath     string `json:"File_Path"`
	FileSize     int    `json:"File_Size"`
	FileType     string `json:"File_Type"`
	ReportPeriod string `json:"Report_Period"`
	ReportType   string `json:"Report_Type"`
	ReportYear   string `json:"Report_Year"`
	NamaEmiten   string `json:"NamaEmiten"`
}

// GetExcelReportLinks -- Return attachments data of type excel
func (fr *FinancialReportAPIResponse) GetExcelAttachments() []Attachment {
	attachments := []Attachment{}
	for _, res := range fr.Results {
		for _, att := range res.Attachments {
			if att.FileType == ".xlsx" {
				attachments = append(attachments, att)
			}
		}
	}
	return attachments
}

// DownloadExcelReports -- Download all available excel reports
func (fr *FinancialReportAPIResponse) DownloadExcelReports() error {
	excelAttachments := fr.GetExcelAttachments()
	for _, att := range excelAttachments {
		directory := filepath.Join("files", "excel_reports", fr.Year, fr.Period)
		err := helpers.Download(directory, att.FileName, att.FilePath)
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}
