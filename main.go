package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
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

func pageMarkup(body, host, yourIP, script string) string {
	return `<!doctype html>
			<html lang="en">
				<head>
					<meta charset="utf-8">
					<title>UnFail2Ban</title>
					<link rel="stylesheet" href="http://` + host + `/UFail2Ban/styles.css">` + script +
		`<meta name="viewport" content="initial-scale=1.0, width=device-width" />
				</head>
				<body>
					<div id='container'>
						<header><h1>UnFail2Ban</h1></header>
						<div><h1><code> Your IP: ` + yourIP + `</code></h1></div>
						` + body + `</div>
						<div class='footer'><footer><small>Website written in Go by Noah Santschi-Cooney<br>This product includes GeoLite2 data created by MaxMind, available from <a href="http://www.maxmind.com">http://www.maxmind.com</a>.</small></footer></div>
					</div>
				</body>
			</html>`
}

func list(ctx *fasthttp.RequestCtx) {
	host, _, err := net.SplitHostPort(string(ctx.Host()[:]))
	if err != nil {
		errorLog.Println(err.Error())
	}

	page := pageMarkup("<div id='table'>"+renderTable(), host, ctx.RemoteAddr().String(), "<script src='http://"+host+"/UFail2Ban/delete.js'></script>")

	ctx.SetContentType("text/html")
	ctx.SetBody([]byte(page))
	ctx.PostBody()
}

func unban(ctx *fasthttp.RequestCtx) {
	//Uncomment the following lines for live
	result := exec.Command("sudo", "fail2ban-client", "set", conf.Jail, "unbanip", strings.TrimPrefix(ctx.QueryArgs().String(), "ip="))
	_, err := result.Output()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
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

func f2bLog(ctx *fasthttp.RequestCtx) {
	host, _, err := net.SplitHostPort(string(ctx.Host()[:]))
	if err != nil {
		errorLog.Println(err.Error())
	}

	page := pageMarkup("<div id='log'>", host, ctx.RemoteAddr().String(), "<script src='http://"+host+"/UFail2Ban/poll.js'></script>")
	ctx.SetContentType("text/html")
	ctx.SetBody([]byte(page))
	ctx.PostBody()
}

func reverse(numbers []string) []string {
	for i := 0; i < len(numbers)/2; i++ {
		j := len(numbers) - i - 1
		numbers[i], numbers[j] = numbers[j], numbers[i]
	}
	return numbers
}

func poll(ctx *fasthttp.RequestCtx) {
	lastDate := strings.TrimSuffix(strings.TrimPrefix(string(ctx.FormValue("date")), "\""), "\"")

	f, err := os.Open("/var/log/fail2ban.log")
	if err != nil {
		ctx.Write([]byte{})
		errorLog.Println(err)
		return
	}
	defer f.Close()

	newLogText, err := ioutil.ReadAll(f)
	if err != nil {
		ctx.Write([]byte{})
		errorLog.Println(err)
		return
	}

	var toSend []string
	splitLog := strings.Split(string(newLogText), "\n")

	if lastDate != "undefined" {
		for i, line := range splitLog {
			if strings.Contains(line, string(lastDate)) {
				toSend = append(toSend, splitLog[i+1:]...)
				break
			}
		}
	} else {
		toSend = splitLog
	}

	ctx.Write([]byte(strings.Join(toSend, "\n")))
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
			list(ctx)
		case "/unban":
			unban(ctx)
		case "/log":
			f2bLog(ctx)
		case "/poll":
			poll(ctx)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	}

	loadConfig()

	info.Println("Fail2Ban jail set to " + conf.Jail)
	info.Println("Listening port set to " + conf.Port)

	fmt.Println("Server started..")

	errorLog.Fatalln(fasthttp.ListenAndServe(":"+conf.Port, requestHandler))
}
