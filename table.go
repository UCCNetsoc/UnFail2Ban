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
	var err error
	if !inDev {
		out, err = exec.Command("iptables", "-L", conf.Jail).Output()
		if err != nil {
			errorLog.Println(err)
			return TableData{}
		}
	} else {
		var err error
		out, err = ioutil.ReadFile("in.txt")
		if err != nil {
			fmt.Print(err)
		}
	}

	rules := func() (ret [][]string) {
		for _, j := range strings.Split(string(out), "\n")[1:] {
			ret = append(ret, strings.Fields(j))
		}
		return
	}()

	rules = filter(rules, func(s string) bool {
		if s == "REJECT" || s == "DROP" {
			return true
		}
		return false
	})

	var tableData TableData

	for _, rule := range rules {
		var row Row
		for j := 0; j < len(rule); {
			if j == 5 {
				row.Data = append(row.Data, strings.Join(rule[j:j+2], " "))
				j += 2
				continue
			}
			row.Data = append(row.Data, rule[j])
			j++
		}

		if !rateLimited {
			rateLimited = getIPInfo(&row, rule[3])
		}
		tableData.Rows = append(tableData.Rows, row)
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
	}...)

	return false
}
