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
	City     string `json:"city"`
	Country  string `json:"country"`
	Region   string `json:"region"`
	Coord    string `json:"loc"`
	Org      string `json:"org"`
	Hostname string `json:"hostname"`
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
			for j := 1; j < len(rules[i]); j++ {
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
	resp, err := http.Get("https://www.ipinfo.io/" + url + "/json")
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
	}

	row.Data = append(row.Data, []string{
		fmt.Sprintf("%s %s %s", ipDetails.City, ipDetails.Region, ipDetails.Country),
		ipDetails.Coord,
		ipDetails.Org,
		ipDetails.Hostname,
	}...)

	return false
}
