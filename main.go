package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
)

type config struct {
	Lang string `toml:"lang"`
	Jail string `toml:"jail"`
	DN   string `toml:"DN"`
	Port string `toml:"port"`

	ExitIfCantLoadLog bool `toml:"exit_if_cant_load_log"`
}

var (
	conf     = &config{}
	info     *log.Logger
	errorLog *log.Logger
	logF     *os.File
)

func fullFatTrim(s []string) (ret []string) {
	for _, i := range s {
		ret = append(ret, " "+strings.Replace(i, " ", "", -1)+" ")
	}
	return
}

func padding(s []string, pad int, padVal string) []string {
	if len(s) >= pad {
		return s
	}

	c := pad - len(s)
	for i := 0; i < c; i++ {
		s = append(s, padVal)
	}
	return s
}

func prepare(s []string) ([][]string, int) {
	var m [][]string
	var pad = 5

	for _, i := range s {
		m = append(m, padding(strings.Fields(i), pad, ""))
	}
	return m, pad
}

func website(ctx *fasthttp.RequestCtx) {
	host, _, err := net.SplitHostPort(string(ctx.Host()[:]))
	if err != nil {
		errorLog.Println(err.Error())
	}

	page := `<!doctype html>
			<html lang="en">
				<head>
					<meta charset="utf-8">
					<title>UnFail2Ban</title>
					<link rel="stylesheet" href="http://` + host + `/styles.css">
					<script src="http://` + host + `/delete.js"></script>
				</head>
				<body>
					<div id='container'>
						<header><h1>UnFail2Ban</h1></header>
						<div><h1><code> Your IP: ` + ctx.RemoteAddr().String() + `</code></h1></div>
						<div id='table'>` + renderTable() + `</div>
						<div class='footer'><footer><small>Website written in Go by Noah Santschi-Cooney<br>This product includes GeoLite2 data created by MaxMind, available from <a href="http://www.maxmind.com">http://www.maxmind.com</a>.</small></footer></div>
					</div>
				</body>
			</html>`

	ctx.SetContentType("text/html")
	ctx.SetBody([]byte(page))
	ctx.PostBody()
}

func unban(ctx *fasthttp.RequestCtx) {
	//Uncomment the following lines for live
	/*  result := exec.Command("sudo", "fail2ban-client", "set", jail, "unbanip", strings.TrimPrefix(ctx.QueryArgs().String(), "ip="))
	 out, err := result.Output()
	if err != nil {
		fmt.Println(err.Error())
		return
	} */
	ctx.Write([]byte(renderTable()))

	info.Println("IP Address", strings.TrimPrefix(ctx.QueryArgs().String(), "ip="), "has been shown mercy")
	return
}

func loadConfig() {
	confRead, err := ioutil.ReadFile("settings.conf")
	if err != nil {
		errorLog.Fatalln("Error reading config file:", err.Error())
	}

	_, err = toml.Decode(string(confRead), conf)
	if err != nil {
		errorLog.Fatalln("Error unmarshalling config:", err.Error())
	}
}

func setLog() *os.File {
	logF, err := os.OpenFile("unf2b.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println(err.Error())
		if conf.ExitIfCantLoadLog {
			os.Exit(2)
		}
	}
	return logF
}

func main() {
	logFile := setLog()
	defer logF.Close()

	log.SetOutput(logFile)

	info = log.New(logFile, "INFO: ", log.Ldate|log.Ltime)
	errorLog = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime)

	info.Println("Starting server...")
	info.Println("Initializing server...")

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

	loadConfig()

	info.Println("Fail2Ban jail set to " + conf.Jail)
	info.Println("Listening port set to " + conf.Port)

	fmt.Println("Server started..")

	err := fasthttp.ListenAndServe(":"+conf.Port, requestHandler)
	if err != nil {
		errorLog.Fatalln(err)
	}
}
