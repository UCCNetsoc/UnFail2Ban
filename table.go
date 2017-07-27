package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
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

// RenderTable generates the HTML table that shows all the entries in the
// Fail2Ban jail specified by the command line argument -jail
func renderTable() (table string) {

	search := "Chain f2b-" + conf.Jail + " (1 references)\n"

	var out []byte
	if !inDev {
		out, _ = exec.Command("iptables", "-L", "-n").Output()
	} else {
		i, _ := os.Open("out.txt")
		defer i.Close()
		out, _ = ioutil.ReadAll(i)
	}

	place := strings.Index(string(out[:]), search)
	cut := out[place+len(search):]
	sep := strings.Split(string(cut[:len(cut)-1]), "\n")[1:]
	ret, pad := prepare(sep)

	table = `<form>
			  <table class="responstable">
			  	  <tr>
					<th>SELECT</th>
					<th>TARGET</th>
					<th>PROT</th>
					<th>OPT</th>
					<th>SOURCE</th>
					<th>DESTINATION</th>
					<th>ADDRESS</th>
					<th>CO-ORDS</th>
					<th>ORGANISATION</th>
					<th>HOST NAME</th>
				  </tr>`

	//Only show IPs that are blocked
	for i := range ret {
		if ret[i][0] == "REJECT" || ret[i][0] == "DROP" {
			table += "<tr class='row'><td><input type='button' class='input' value='Unban'></input></td>"
			for j := 0; j < pad; j++ {
				table += "<td>" + ret[i][j] + "</td>"
			}
			resp, err := http.Get("https://www.ipinfo.io/" + ret[i][3] + "/json")
			if err != nil {
				errorLog.Println(err)
				break
			} else if resp.StatusCode == 429 {
				table = "<p class='error'>To many requests to https://ipinfo.io/ <br>Rate limit is 1000 requests per day. Please contact a Sys Admin about this or read the error log if you are one.</p>"
				break
			}
			defer resp.Body.Close()

			var ipDetails ipInfo
			ipData, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				errorLog.Println(err)
			}
			json.Unmarshal(ipData, &ipDetails)

			ipinfo := []string{ipDetails.City, ipDetails.Region, ipDetails.Country}
			table += "<td>" + strings.Join(ipinfo, ", ") + "</td>"
			table += "<td>" + ipDetails.Coord + "</td>"
			table += "<td>" + ipDetails.Org + "</td>"
			table += "<td>" + ipDetails.Hostname + "</td>"
			table += "</tr>" 

		}
	}
	table += "</table></form>"
	return
}
