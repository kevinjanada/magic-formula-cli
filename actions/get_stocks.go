package actions

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	"github.com/kevinjanada/magic-formula-cli/helpers"
	"github.com/kevinjanada/magic-formula-cli/models"
	"github.com/urfave/cli"
)

func GetStocks(db *gorm.DB) func(c *cli.Context) {
	return func(c *cli.Context) {
		stocks, err := FetchStocks()
		if err != nil {
			panic(err)
		}

		err = SaveStockDataToDB(stocks, db)
		if err != nil {
			panic(err)
		}
	}
}

type StockData struct {
	Code         string `json:"Code"`
	Name         string `json:"Name"`
	ListingDate  string `json:"ListingDate"`
	Shares       int64  `json:"Shares"`
	ListingBoard string `json:"ListingBoard"`
	Links        []Link `json:"Links"`
}

type Link struct {
	Rel    string `json:"Rel"`
	Href   string `json:"Href"`
	Method string `json:"Method"`
}

type StockAPIResponse struct {
	Draw            int         `json:"draw"`
	RecordsTotal    int         `json:"recordsTotal"`
	RecordsFiltered int         `json:"recordsFiltered"`
	Data            []StockData `json:"data"`
}

// generateFetchStockURL --
func generateFetchStockURL(start int, length int) string {
	return fmt.Sprintf(
		`https://www.idx.co.id/umbraco/Surface/StockData/GetSecuritiesStock?start=%d&length=%d`,
		start,
		length,
	)
}

// FetchStocks -- Fetch Stocks Data from IDX API
func FetchStocks() ([]StockData, error) {
	start := 0
	length := 10
	URL := generateFetchStockURL(start, length)
	resp, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	aggregatedResponse := &StockAPIResponse{}
	helpers.JSONToStruct(resp, aggregatedResponse)

	numOfStocksLeft := aggregatedResponse.RecordsTotal
	start += length
	for numOfStocksLeft > 0 {
		URL := generateFetchStockURL(start, length)
		resp, err := http.Get(URL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		nextResponse := &StockAPIResponse{}
		helpers.JSONToStruct(resp, nextResponse)

		aggregatedResponse.Data = append(aggregatedResponse.Data, nextResponse.Data...)

		start += length
		numOfStocksLeft -= length
	}

	return aggregatedResponse.Data, nil
}

// SaveStockDataToDB -- Receive Stock Data and save them all to database
func SaveStockDataToDB(stocksData []StockData, db *gorm.DB) error {
	for _, sd := range stocksData {
		stockModel := &models.Stock{}
		query := db.Where("code = ?", sd.Code)
		query.First(stockModel)
		// If stock exists, update
		if stockModel.ID != uuid.Nil {
			stockModel.Shares = sd.Shares
			stockModel.ListingBoard = sd.ListingBoard
		} else { // Else Create
			layout := "2006-01-02T15:04:05"
			parsedDate, err := time.Parse(layout, sd.ListingDate)
			if err != nil {
				fmt.Println(err)
			}
			stockModel = &models.Stock{
				Code:         sd.Code,
				Name:         sd.Name,
				ListingDate:  parsedDate,
				Shares:       sd.Shares,
				ListingBoard: sd.ListingBoard,
			}
		}
		db.Save(stockModel)
	}
	return nil
}
