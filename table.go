package main

import (
	"encoding/json"
	"fmt"
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

func renderTable() (tableData TableData) {
	var rateLimited bool

	//TODO use fail2rest
	out, err := exec.Command("iptables", "-L", conf.Jail).Output()
	if err != nil {
		errorLog.Println(err)
		return
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

	for _, rule := range rules {
		var row Row
		var extendRow [4]string

		for j := 0; j < len(rule); j++ {
			if j == 5 {
				row.Data = append(row.Data, strings.Join(rule[j:j+2], " "))
				j++
				continue
			}
			row.Data = append(row.Data, rule[j])
		}

		if !rateLimited {
			// Gonna move this to client side later
			// By doing this, we spread the requests over multiple IPs rather than
			// all of them originating from our servers.
			// Or not, will see
			extendRow, rateLimited = getIPInfo(rule[3])
		}
		row.Data = append(row.Data, extendRow[:]...)

		tableData.Rows = append(tableData.Rows, row)
	}

	tableData.NotEmpty = len(tableData.Rows) > 0

	return
}

func getIPInfo(url string) ([4]string, bool) {
	resp, err := http.Get("http://ip-api.com/json/" + url)
	if err != nil {
		errorLog.Println(err)
		return [4]string{}, false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return [4]string{"Ratelimited", "By", "ipinfo.io", ":("}, true
	}

	var ipDetails ipInfo
	if err := json.NewDecoder(resp.Body).Decode(&ipDetails); err != nil {
		errorLog.Println(err)
		return [4]string{}, false
	}
	if ipDetails.Status != "success" && ipDetails.Message == "over quota" {
		return [4]string{}, true
	}

	return [4]string{fmt.Sprintf("%s %s %s", ipDetails.City, ipDetails.Region, ipDetails.Country), fmt.Sprintf("Lat: %f Lon: %f", ipDetails.Lat, ipDetails.Lon), ipDetails.Org}, false
}
