package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type config struct {
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

	inDev = true
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

func pageMarkup(body, script string, r *http.Request) string {
	showIP := "<div id='ip'><h2><code> Your IP: " + r.RemoteAddr + "</code></h2></div>"

	return `
<!doctype html>
<html lang="en">
	<head>
		<meta charset="utf-8">
		<title>UnFail2Ban</title>
		<link rel="stylesheet" href="http://localhost/UFail2Ban/styles.css">` +
		script +
		`<meta name="viewport" content="initial-scale=1.0, width=device-width" />
	</head>
	<body>
		<header>
			<span>
				<h1>UnFail2Ban</h1>
				<a href='/log'>Log</a>
				<a href='/list'>List</a>
			</span>
			` +
		showIP +
		`</header>
		<main>` +
		body +
		`<footer><small>Website written in Go by Noah Santschi-Cooney<br>This product includes GeoLite2 data created by MaxMind, available from <a href="http://www.maxmind.com">http://www.maxmind.com</a>.</small></footer>
		</main>
	</body>
</html>`
}

func list(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	page := pageMarkup("<div id='table'>"+renderTable()+"</div>", "<script src='http://localhost/UFail2Ban/delete.js'></script>", r)

	fmt.Fprint(w, page)
}

func unban(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if !inDev {
		result := exec.Command("sudo", "fail2ban-client", "set", conf.Jail, "unbanip", r.URL.Query()["ip"][0])
		_, err := result.Output()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	fmt.Fprint(w, renderTable())

	info.Println("IP Address", r.URL.Query()["ip"][0], "has been shown mercy")
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

func f2bLog(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	page := pageMarkup("<div id='log'></div>", "<script src='http://localhost/UFail2Ban/poll.js'></script>", r)
	fmt.Fprint(w, page)
}

func reverse(numbers []string) []string {
	for i := 0; i < len(numbers)/2; i++ {
		j := len(numbers) - i - 1
		numbers[i], numbers[j] = numbers[j], numbers[i]
	}
	return numbers
}

func poll(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	lastDate := strings.TrimSuffix(strings.TrimPrefix(r.URL.Query()["date"][0], "\""), "\"")

	f, err := os.Open("/var/log/fail2ban.log")
	if err != nil {
		fmt.Fprint(w)
		errorLog.Println(err)
		return
	}
	defer f.Close()

	newLogText, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Fprint(w, "")
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

	fmt.Fprint(w, strings.Join(toSend, "\n"))
}

func notFound(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, pageMarkup("<div style='font-size: 2em;'>404 Page doesn't exist</div>", "", r))
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, pageMarkup("<p style='font-size: 3em'>UnFail2Ban</p><br>\n"+
		"<p>Your one-stop web GUI for Fail2Ban administration and monitoring</p>", "", r))
}

func main() {
	logFile := setLog()
	defer logF.Close()

	log.SetOutput(logFile)

	info = log.New(logFile, "INFO: ", log.Ldate|log.Ltime)
	errorLog = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime)

	info.Println("Starting server...")
	info.Println("Initializing server...")

	r := mux.NewRouter()

	r.HandleFunc("/home", home)
	r.HandleFunc("/list", list)
	r.HandleFunc("/unban", unban)
	r.HandleFunc("/log", f2bLog)
	r.HandleFunc("/poll", poll)
	r.HandleFunc("/", notFound)

	loadConfig()

	info.Println("Fail2Ban jail set to " + conf.Jail)
	info.Println("Listening port set to " + conf.Port)

	fmt.Println("Server started..")

	errorLog.Fatalln(http.ListenAndServe(":"+conf.Port, r))
}
