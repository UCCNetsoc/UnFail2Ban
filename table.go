package main

import (
	"github.com/oschwald/geoip2-golang"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

// RenderTable generates the HTML table that shows all the entries in the
// Fail2Ban jail specified by the command line argument -jail
func renderTable() (table string) {

	search := "Chain f2b-" + conf.Jail + " (1 references)\n"
	//Uncomment the next line for live
	//out, _ := exec.Command("iptables", "-L", "-n").Output()
	//Uncomment the following for testing
	i, _ := os.Open("out.txt")
	defer i.Close()
	out, _ := ioutil.ReadAll(i)

	place := strings.Index(string(out[:]), search)
	cut := out[place+len(search):]
	sep := strings.Split(string(cut[:len(cut)-1]), "\n")[1:]
	ret, pad := prepare(sep)

	db, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		errorLog.Println("Open GeoLite2-City.mmdb error")
		return "<p class='error'>Error. Please contact your Sys Admin or read the error log if you are one.</p><code class='error'>" + err.Error() + "</code>"
	}
	defer db.Close()

	table = `<form>
			  <table class="responstable">
			  	  <tr>
					  <th>SELECT</th>
					  <th>TARGET</th>
					  <th>PROT</th>
					  <th>OPT</th>
					  <th>SOURCE</th>
					  <th>DESTINATION</th>
				  	  <th>COUNTRY</th>
				  </tr>`

	//Only show IPs that are blocked
	for i := range ret {
		if ret[i][0] == "REJECT" {
			table += "<tr class='row'><td><input type='button' class='input' value='Unban'></input></td>"
			for j := 0; j < pad; j++ {
				table += "<td>" + ret[i][j] + "</td>"
			}

			ip := net.ParseIP(ret[i][3])
			record, err := db.City(ip)
			if err != nil {
				errorLog.Println(err.Error())
			}
			table += "<td>" + record.Country.Names[conf.Lang] + ", " + record.City.Names[conf.Lang] + "</td>"
			table += "</tr>"
		}
	}
	table += "</table></form>"
	return
}
