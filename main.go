package main

import (
	"strings"
	"fmt"
	"os"
	"log"
	"net"
	"github.com/valyala/fasthttp"
	"encoding/json"
	"io/ioutil"
)

type config struct {
	Lang string `json:"lang"`
	Jail string `json:"jail"`
	DN   string `json:"DN"`
}

var conf = &config{}
var info *log.Logger
var error *log.Logger

/*func init(){
	flag.StringVar(&jail, "jail", "", "Jail")
	flag.StringVar(&lang, "lang", "", "Language of Countries")
	flag.Parse()

	if jail == "" {
		fmt.Println("Fail2Ban Jail must be supplied. Example: -jail myJail")
		os.Exit(0)
	}


}*/

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

func website(ctx *fasthttp.RequestCtx){
	host, _, err := net.SplitHostPort(string(ctx.Host()[:])); if err != nil { error.Println(err.Error()) }
	page := `<!doctype html>
			<html lang="en">
				<head>
					<meta charset="utf-8">
					<title>UnFail2Ban</title>
					<link rel="stylesheet" href="http://`+host+`/styles.css">
					<script src="http://`+host+`/delete.js"></script>
				</head>
				<body>
					<div id='container'>
						<div><h1>UnFail2Ban</h1></div>
						<div><h1><code>`+ctx.RemoteAddr().String()+`</code></h1></div>
						<div id='table'>`+RenderTable()+`</div>
						<div class='footer'><footer><small>Website written in Go by Noah Santschi-Cooney<br>This product includes GeoLite2 data created by MaxMind, available from <a href="http://www.maxmind.com">http://www.maxmind.com</a>.</small></footer></div>
					</div>
				</body>
			</html>`
	ctx.SetContentType("text/html")
	ctx.SetBody([]byte(page))
	ctx.PostBody()
}

func unban(ctx *fasthttp.RequestCtx){
	//Uncomment the following lines for live
	// result := exec.Command("sudo", "fail2ban-client", "set", jail, "unbanip", strings.TrimPrefix(ctx.QueryArgs().String(), "ip="))
	// out, err := result.Output(); if err != nil { fmt.Println(err.Error()); return }
	// fmt.Println(err.Error(), out)
	ctx.Write([]byte(RenderTable()))

	info.Println("IP Address", strings.TrimPrefix(ctx.QueryArgs().String(), "ip="), "has been shown mercy")
	return
}

func loadConfig(){
	confRead, err := ioutil.ReadFile("unf2b.json")
	if err != nil { 
		error.Fatalln("Error reading config file:", err.Error())
	}
	
	err = json.Unmarshal(confRead, conf)
	if err != nil {
		error.Fatalln("Error unmarshalling config:", err.Error())
	}
}

func main(){
	requestHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/list":
			website(ctx)
		case "/unban":
			unban(ctx)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	}

	logF, err := os.OpenFile("unf2b.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644); if err != nil { fmt.Println(err.Error()); }
	defer logF.Close()

	log.SetOutput(logF)

	info  = log.New(logF, "INFO: ", log.Ldate | log.Ltime)
	error = log.New(logF, "ERROR: ", log.Ldate | log.Ltime)

	loadConfig()

	info.Println("Initializing server. Fail2Ban jail set to `"+conf.Jail+"`")
	fmt.Println("Server started..")

	err = fasthttp.ListenAndServe(":8080", requestHandler)
}