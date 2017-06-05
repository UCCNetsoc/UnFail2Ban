package main 

import (
	"os"
	"io/ioutil"
	"strings"
	"github.com/oschwald/geoip2-golang"
	"strconv"
	"net"
)

func renderTable() string {
	search := "Chain f2b-"+jail+" (1 references)\n"
	//out, _ := exec.Command("iptables", "-L", "-n").Output()
	i,_ := os.Open("out.txt")
	defer i.Close()
	out,_ := ioutil.ReadAll(i)
	place := strings.Index(string(out[:]), search)
	cut := out[place+len(search):]
	sep := strings.Split(string(cut[:len(cut)-1]), "\n")[1:]
	ret, pad := prepare(sep)
	
	db, err := geoip2.Open("GeoLite2-City.mmdb")
	defer db.Close()
    if err != nil {
        error.Println("Open GeoLite2-City.mmdb error")
		return "<p>Error. Please contact your Sys Admin or read the error log if you are one.</p>"
    }

	table := `<form>
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

	for i := range ret {
		if ret[i][0] == "REJECT" {
			table += "<tr class='row'><td><input type='radio' name='input' id='input' value="+strconv.Itoa(i)+"></input></td>"
			for j := 0; j < pad; j++ {
					table += "<td>"+ret[i][j]+"</td>"
			}
			    // If you are using strings that may be invalid, check that ip is not nil
			ip := net.ParseIP(ret[i][3])
			record, err := db.City(ip)
			if err != nil {
				error.Println(err.Error())
			}
			table += "<td>"+record.Country.Names[lang]+", "+record.City.Names[lang]+"</td>"
			table += "</tr>"
		}
	}
	table += "</table><input type='submit' id='submit' value='Unban'></input></form>"

	return table
}