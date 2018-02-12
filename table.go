package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
)

type ipInfo struct {
	Status  string  `json:"status"`
	City    string  `json:"city"`
	Country string  `json:"countryCode"`
	Region  string  `json:"region"`
	Lat     float32 `json:"lat"`
	Lon     float32 `json:"lon"`
	Org     string  `json:"org"`
	Message string  `json:"message"`
}

type Row struct {
	Data []string
}

type TableData struct {
	IP       string
	NotEmpty bool
	Rows     []Row
}

// RenderTable generates the HTML table that shows all the entries in the
// Fail2Ban jail specified by the command line argument -jail
func renderTable() TableData {
	var out []byte
	var rateLimited bool

	if !inDev {
		out, _ = exec.Command("iptables", "-L", conf.Jail).Output()
	} else {
		var err error
		out, err = ioutil.ReadFile("in.txt")
		if err != nil {
			fmt.Print(err)
		}
	}

	rules := func() [][]string {
		in := strings.Split(string(out), "\n")[1:]
		out := make([][]string, len(in))
		for i, j := range in {
			out[i] = strings.Fields(j)
		}
		return out
	}()

	tableData := TableData{
		Rows: make([]Row, 0),
	}

	for i := range rules {
		if rules[i][0] == "REJECT" || rules[i][0] == "DROP" {
			var row Row
			for j := 0; j < len(rules[i])-2; j++ {
				row.Data = append(row.Data, rules[i][j])
			}

			if !rateLimited {
				rateLimited = getIPInfo(&row, rules[i][3])
			}
			tableData.Rows = append(tableData.Rows, row)
		}
	}

	tableData.NotEmpty = len(tableData.Rows) > 0

	return tableData
}

func getIPInfo(row *Row, url string) bool {
	resp, err := http.Get("http://ip-api.com/json/" + url)
	if err != nil {
		errorLog.Println(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 429 {
		//tableData.Rows = "<p class='error'>To many requests to https://ipinfo.io/ <br>Rate limit is 1000 requests per day. Please contact a Sys Admin about this or read the error log if you are one.</p>"
		row.Data = append(row.Data, []string{
			"Ratelimited",
			"By",
			"ipinfo.io",
			":(",
		}...)
		return true
	}

	var ipDetails ipInfo
	if err := json.NewDecoder(resp.Body).Decode(&ipDetails); err != nil {
		errorLog.Println(err)
		return false
	}
	if ipDetails.Status != "success" && ipDetails.Message == "over quota" {
		return true
	}

	row.Data = append(row.Data, []string{
		fmt.Sprintf("%s %s %s", ipDetails.City, ipDetails.Region, ipDetails.Country),
		fmt.Sprintf("Lat: %f Lon: %f", ipDetails.Lat, ipDetails.Lon),
		ipDetails.Org,
	}...)

	return false
}
