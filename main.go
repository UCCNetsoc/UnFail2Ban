package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"html/template"
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

func list(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data := renderTable()
	data.IP = r.RemoteAddr

	tmpl, err := template.ParseFiles("static/main.html", "static/table.html")
	if err != nil {
		fmt.Println(err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.ExecuteTemplate(w, "main", data)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func unban(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	fmt.Fprint(w, r.URL.Query()["ip"][0])
	return
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
}

func loadConfig() {
	confRead, err := ioutil.ReadFile("settings.conf")
	if err != nil {
		errorLog.Fatalln("Error reading config file:", err)
	}

	_, err = toml.Decode(string(confRead), conf)
	if err != nil {
		errorLog.Fatalln("Error unmarshalling config:", err)
	}
}

func setLog() *os.File {
	logF, err := os.OpenFile("unf2b.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println(err.Error())
		if conf.ExitIfCantLoadLog {
			log.Fatalln("no log provided")
		}
	}
	return logF
}

func f2bLog(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	//page := pageMarkup("<div id='log'></div>", "<script src='http://localhost/UFail2Ban/poll.js'></script>", r)
	fmt.Fprint(w, nil)
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

func main() {
	logFile := setLog()
	defer logF.Close()

	log.SetOutput(logFile)

	info = log.New(logFile, "INFO: ", log.Ldate|log.Ltime)
	errorLog = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime)

	info.Println("Starting server...")
	info.Println("Initializing server...")

	//http.HandleFunc("/home", home)
	http.HandleFunc("/list", list)
	http.HandleFunc("/unban", unban)
	http.HandleFunc("/log", f2bLog)
	http.HandleFunc("/poll", poll)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	loadConfig()

	info.Println("Fail2Ban jail set to " + conf.Jail)
	info.Println("Listening port set to " + conf.Port)

	fmt.Println("Server started..\nListening on http://127.0.0.1:" + conf.Port)

	errorLog.Fatalln(http.ListenAndServe(":"+conf.Port, nil))
}
