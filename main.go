package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/go-chi/chi"
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
	result := exec.Command("sudo", "fail2ban-client", "set", strings.TrimPrefix(conf.Jail, "f2b-"), "unbanip", r.URL.Query()["ip"][0])
	var b bytes.Buffer
	result.Stderr = &b
	if _, err := result.Output(); err != nil {
		w.WriteHeader(500)
		errorLog.Println(b.String())
		return
	}

	info.Println("IP Address", r.URL.Query()["ip"][0], "has been shown mercy")
}

func f2bLog(w http.ResponseWriter, r *http.Request) {
	//page := pageMarkup("<div id='log'></div>", "<script src='http://localhost/UFail2Ban/poll.js'></script>", r)
	fmt.Fprint(w, nil)
}

func poll(w http.ResponseWriter, r *http.Request) {
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

	user, err := getUserFromLDAP(username, password)
	if err != nil {
		switch {
		case err == errWrongPass || err == errNoUser:
			renderLogin(w, r, "Username or password was incorrect")
			errorLog.Println(fmt.Sprintf("IP %s failed login with username %s", r.RemoteAddr, username))
		default:
			renderLogin(w, r, "Error processing your request")
			errorLog.Println(err)
		}
	}

	if !user.isadmin {
		notAuthorized(w, r)
		return
	}

	http.Redirect(w, r, "/list", http.StatusFound)
}

func notAuthorized(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("static/main.html", "static/noauth.html")
	if err != nil {
		fmt.Println(err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.ExecuteTemplate(w, "main", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func main() {
	info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime)

	info.Println("Starting server...")
	info.Println("Initializing server...")

	r := chi.NewRouter()

	r.HandleFunc("/", home)

	//auth group. cookie middleware to be added
	r.Group(func(r chi.Router) {
		r.Get("/list", list)
		r.Delete("/unban", unban)
		r.Get("/poll", poll)
	})

	r.Post("/login", login)

	r.Mount("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	loadConfig()

	info.Println("Fail2Ban jail set to " + conf.Jail)
	info.Println("Listening port set to " + conf.Port)

	fmt.Println("Server started..\nListening on http://127.0.0.1:" + conf.Port)
	errorLog.Fatalln(http.ListenAndServe(":"+conf.Port, r))
}
