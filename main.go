package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/go-chi/chi"
	"gopkg.in/ldap.v2"
)

type config struct {
	Jail string `toml:"jail"`
	Port string `toml:"port"`

	LDAPKey    string `toml:"LDAP_Key"`
	LDAPHost   string `toml:"LDAP_Host"`
	LDAPUser   string `toml:"LDAP_User"`
	LDAPBaseDN string `toml:"LDAP_BaseDN"`
}

var (
	conf     = &config{}
	info     *log.Logger
	errorLog *log.Logger
	logF     *os.File
)

func list(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data := struct {
		Data TableData
		IP   string
	}{
		renderTable(),
		r.RemoteAddr,
	}

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

	result := exec.Command("sudo", "fail2ban-client", "set", conf.Jail, "unbanip", r.URL.Query()["ip"][0])
	_, err := result.Output()
	if err != nil {
		fmt.Println(err.Error())
		return
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
		log.Fatalln("no log provided")
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

func home(w http.ResponseWriter, r *http.Request) {
	renderLogin(w, r, "")
}

func renderLogin(w http.ResponseWriter, r *http.Request, msg string) {
	tmpl, err := template.ParseFiles("static/main.html", "static/form.html")
	if err != nil {
		fmt.Println(err)
		return
	}

	data := struct {
		Data string
	}{
		msg,
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.ExecuteTemplate(w, "main", data)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	l, err := ldap.Dial("tcp", conf.LDAPHost)
	if err != nil {
		errorLog.Println(err)
		return
	}
	defer l.Close()

	if err = l.Bind(conf.LDAPUser, conf.LDAPKey); err != nil {
		errorLog.Println(err)
		return
	}

	searchRequest := ldap.NewSearchRequest(
		conf.LDAPBaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		fmt.Sprintf("(&(objectClass=account)(uid=%s))",
			ldap.EscapeFilter(username)),
		[]string{},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		errorLog.Println(err)
		return
	}

	if len(sr.Entries) != 1 {
		fmt.Fprint(w, "User does not exist")
		return
	}

	if err := l.Bind(sr.Entries[0].DN, password); err != nil {
		renderLogin(w, r, "Wrong password or username")
		return
	}
	fmt.Fprint(w, "Successful login")
}

func main() {
	logFile := setLog()
	defer logF.Close()

	log.SetOutput(logFile)

	info = log.New(logFile, "INFO: ", log.Ldate|log.Ltime)
	errorLog = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime)

	info.Println("Starting server...")
	info.Println("Initializing server...")

	r := chi.NewRouter()

	r.HandleFunc("/", home)
	r.HandleFunc("/list", list)
	r.HandleFunc("/unban", unban)
	r.HandleFunc("/log", f2bLog)
	r.HandleFunc("/poll", poll)
	r.Post("/login", login)
	r.Mount("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	loadConfig()

	info.Println("Fail2Ban jail set to " + conf.Jail)
	info.Println("Listening port set to " + conf.Port)

	fmt.Println("Server started..\nListening on http://127.0.0.1:" + conf.Port)

	errorLog.Fatalln(http.ListenAndServe(":"+conf.Port, r))
}
