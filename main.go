package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
	"github.com/kevinjanada/magic-formula-cli/actions"
	"github.com/urfave/cli"
)

func initEnv() {
	if err := godotenv.Load(); err != nil {
		log.Print(err)
	}
}

func connectDB() (*gorm.DB, error) {
	postgresHost, _ := os.LookupEnv("POSTGRES_HOST")
	postgresPort, _ := os.LookupEnv("POSTGRES_PORT")
	postgresDB, _ := os.LookupEnv("POSTGRES_DB")
	postgresUser, _ := os.LookupEnv("POSTGRES_USER")
	postgresPassword, _ := os.LookupEnv("POSTGRES_PASSWORD")
	db, err := gorm.Open(
		"postgres",
		fmt.Sprintf(
			"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
			postgresHost,
			postgresPort,
			postgresDB,
			postgresUser,
			postgresPassword,
		),
	)
	if err != nil {
		return nil, err
	}
	return db, nil
}

var app = cli.NewApp()

func info() {
	app.Name = "Magic Formula CLI"
}

func commands(db *gorm.DB) {
	app.Commands = []cli.Command{
		{
			Name:   "get-stocks",
			Usage:  "Get stock data and save them to DB",
			Action: actions.GetStocks(db),
		},
		{
			Name:   "download-reports",
			Usage:  "Download excel reports",
			Action: actions.DownloadReports(db),
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "year, y",
					Value: "2019",
					Usage: "Get reports for the year",
				},
				cli.StringFlag{
					Name:  "period, p",
					Value: "3",
					Usage: "Period of the year",
				},
			},
		},
		{
			Name:   "import-reports",
			Usage:  "Import excel reports to DB",
			Action: actions.ImportReports(db),
		},
	}
}

func main() {
	initEnv()
	info()

	db, err := connectDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	commands(db)
	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
