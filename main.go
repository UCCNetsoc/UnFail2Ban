package main

import (
	"strings"
	"net/http"
	"fmt"
	"os"
	"flag"
	"log"
)

var jail string
var lang string
var info *log.Logger
var error *log.Logger

func init(){
	flag.StringVar(&jail, "jail", "", "Jail")
	flag.StringVar(&lang, "lang", "en", "Language of Countries")
	flag.Parse()

	if jail == "" {
		fmt.Println("Fail2Ban Jail must be supplied. Example: -jail myJail")
		os.Exit(0)
	}
}

func fullFatTrim(s []string) (ret []string){
	for _,i := range s {
		ret = append(ret, " "+strings.Replace(i, " ", "", -1)+" ")
	}
	return
}

func padding(s []string, pad int, padVal string) []string{
	if len(s) >= pad {
		return s
	}
	c := pad-len(s)
	for i := 0; i < c; i++ {
		s = append(s, padVal)
	}
	return s
}

func prepare(s []string) ([][]string, int){
	var m [][]string
	pad := 5
	for _,i := range s {
		m = append(m, padding(strings.Fields(i),pad, ""))
	}
	return m, pad
}

func website(w http.ResponseWriter, r *http.Request){
	page := `<!doctype html>
			<html lang="en">
				<head>
					<meta charset="utf-8">
					<title>UnFail2Ban</title>
					<link rel="stylesheet" href="http://127.0.0.1/styles.css">
					<script src="http://127.0.0.1/delete.js"></script>
				</head>
				<body>
					<div id='container'>
						<div><h1>UnFail2Ban</h1></div>
						<div><h1><code>`+r.RemoteAddr+`</code></h1></div>
						<div id='table'>`+renderTable()+`</div>
						<div class='footer'><footer><small>Website written in Go by Noah Santschi-Cooney<br>This product includes GeoLite2 data created by MaxMind, available from <a href="http://www.maxmind.com">http://www.maxmind.com</a>.</small></footer></div>
					</div>
				</body>
			</html>`


	w.Write([]byte(page))
}

func unban(w http.ResponseWriter, r *http.Request){
	// result := exec.Command("sudo", "fail2ban-client", "set", jail, "unbanip", r.URL.Query()["ip"][0])
	// out, err := result.Output(); if err != nil { fmt.Println(err.Error()); return }
	// fmt.Println(err.Error(), out)
	w.Write([]byte(renderTable()))

	info.Println("IP Address", r.URL.Query()["ip"][0], "has been shown mercy")
	return
}

func main(){
	http.HandleFunc("/list", website)
	http.HandleFunc("/unban", unban)

	f, err := os.OpenFile("unf2b.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644); if err != nil { fmt.Println(err.Error()) }
	defer f.Close()

	log.SetOutput(f)

	info  = log.New(f, "INFO: ", log.Ldate | log.Ltime)
	error = log.New(f, "ERROR: ", log.Ldate | log.Ltime)

	info.Println("Initializing server. Fail2Ban jail set to `"+jail+"`")

	fmt.Println("Server started..")

	err = http.ListenAndServe(":8080", nil); if err != nil { fmt.Println(err.Error()) }
}