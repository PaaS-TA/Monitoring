package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/influxdata/influxdb/client"
)

func main() {
	influxUrl := flag.String("influxUrl", "http://localhost:8086", "influx endpoint")
	influxUser := flag.String("user", "root", "influx user")
	influxPassword := flag.String("password", "root", "influx user's password")
	databaseName := flag.String("database", "", "Database name to create")
	retention := flag.String("retention", "14d", "Retention duration")
	replication := flag.String("replication", "1", "Replication count")
	flag.Parse()

	if len(*databaseName) == 0 {
		fmt.Println("Database name is required")
		flag.Usage()
		os.Exit(1)
	}

	influxServerUrl, _ := url.Parse(*influxUrl)
	influxConfig := client.Config{
		URL:       *influxServerUrl,
		Username:  *influxUser,
		Password:  *influxPassword,
		UserAgent: "influx-bootstrap",
	}
	dbClient, err := client.NewClient(influxConfig)
	if err != nil {
		panic(fmt.Sprintf("Error connecting to Influx DB: %s", err))
	}

	_, err = dbClient.Query(client.Query{Command: fmt.Sprintf("create database %s", *databaseName)})
	if err != nil {
		panic(fmt.Sprintf("Error creating DB '%s': %s", *databaseName, err))
	}

	retentionCommand := fmt.Sprintf("create RETENTION POLICY \"default\" ON %s DURATION %s REPLICATION %s DEFAULT", *databaseName, *retention, *replication)
	_, err = dbClient.Query(client.Query{Command: retentionCommand})
	if err != nil {
		panic(fmt.Sprintf("Error creating default retention policy: %s", err))
	}
}
