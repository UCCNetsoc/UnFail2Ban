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
	"github.com/gorilla/context"
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
	conf     = new(config)
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
		w.WriteHeader(http.StatusInternalServerError)
		errorLog.Println(err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.ExecuteTemplate(w, "main", data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorLog.Println(err)
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
	val := r.Header.Get("auth")
	switch val {
	case "":
		renderLogin(w, r, "")
	case "njet":
		renderLogin(w, r, "Username or password was incorrect")
	case "err":
		renderLogin(w, r, "Error processing your request")
	case "nologin":
		renderLogin(w, r, "Please log in to view this page")
	}
}

func renderLogin(w http.ResponseWriter, r *http.Request, msg string) {
	tmpl, err := template.ParseFiles("static/main.html", "static/form.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorLog.Println(err)
		return
	}

	data := struct {
		Data string
	}{
		msg,
	}

	w.Header().Set("Content-Type", "text/html")
	if err = tmpl.ExecuteTemplate(w, "main", data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorLog.Println(err)
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
			w.Header().Set("auth", "njet")
			errorLog.Println(fmt.Sprintf("IP %s failed login with username %s", r.RemoteAddr, username))
		default:
			w.Header().Set("auth", "err")
			errorLog.Println(err)
		}
	}

	if !user.isadmin {
		http.Redirect(w, r, "/noauth", http.StatusUnauthorized)
		return
	}

	http.Redirect(w, r, "/list", http.StatusFound)
}

func notAuthorized(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("static/main.html", "static/noauth.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorLog.Println(err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err = tmpl.ExecuteTemplate(w, "main", nil); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorLog.Println(err)
	}
}

func main() {
	info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime)

	info.Println("Starting server...")
	info.Println("Initializing server...")

	r := chi.NewRouter()

	r.HandleFunc("/", home)
	r.HandleFunc("/noauth", notAuthorized)

	//auth group. cookie middleware to be added
	r.Group(func(r chi.Router) {
		r.Use(checkCookie)
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
	errorLog.Fatalln(http.ListenAndServe(":"+conf.Port, context.ClearHandler(r)))
}
