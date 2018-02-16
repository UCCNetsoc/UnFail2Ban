package main

import (
	"bytes"
	"encoding/gob"
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
	"github.com/gorilla/sessions"
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
	infoLog  *log.Logger
	errorLog *log.Logger
	logF     *os.File
	store    = sessions.NewCookieStore([]byte(conf.LDAPKey))
)

func init() {
	store.Options = &sessions.Options{
		Domain:   "localhost",
		MaxAge:   60 * 10,
		HttpOnly: true,
	}

	gob.Register(user{})
}

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
	u, err := getUserFromSession(r)
	if err != nil {
		//to be modified to handle different errors
		w.WriteHeader(http.StatusForbidden)
		return
	}

	result := exec.Command("sudo", "fail2ban-client", "set", strings.TrimPrefix(conf.Jail, "f2b-"), "unbanip", r.URL.Query()["ip"][0])
	var b bytes.Buffer
	result.Stderr = &b
	if _, err := result.Output(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorLog.Println(b.String())
		return
	}

	infoLog.Println(fmt.Sprintf("IP Address %s has been unbanned by %s", r.URL.Query()["ip"][0], u.Username))
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
	val := r.URL.Query().Get("auth")

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

	infoLog.Println(fmt.Sprintf("login request from %s for username %s", r.RemoteAddr, username))

	user, err := getUserFromLDAP(username, password)
	if err != nil {
		switch {
		case err == errWrongPass || err == errNoUser:
			errorLog.Println(fmt.Sprintf("IP %s failed login with username %s", r.RemoteAddr, username))
			http.Redirect(w, r, "/?auth=njet", http.StatusTemporaryRedirect)
			return
		default:
			errorLog.Println(err)
			http.Redirect(w, r, "/?auth=err", http.StatusTemporaryRedirect)
			return
		}
	}

	if !user.isadmin {
		http.Redirect(w, r, "/noauth", http.StatusTemporaryRedirect)
		infoLog.Println(fmt.Sprintf("non-admin %s attempted login from %s", username, r.RemoteAddr))
		return
	}

	session, err := store.New(r, "id")
	if err != nil {
		http.Redirect(w, r, "/?auth=err", http.StatusTemporaryRedirect)
		errorLog.Println("session error", err)
		return
	}

	session.Values["user"] = user

	if err := session.Save(r, w); err != nil {
		http.Redirect(w, r, "/?auth=err", http.StatusTemporaryRedirect)
		errorLog.Println("error saving session", err)
		return
	}

	infoLog.Println(fmt.Sprintf("%s successfully logged in from %s", username, r.RemoteAddr))

	http.Redirect(w, r, "/list", http.StatusTemporaryRedirect)
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
	infoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime)

	infoLog.Println("Starting server...")
	infoLog.Println("Initializing server...")

	r := chi.NewRouter()

	r.HandleFunc("/", home)
	r.HandleFunc("/noauth", notAuthorized)

	//auth group.
	r.Group(func(r chi.Router) {
		r.Use(checkCookie)
		r.HandleFunc("/list", list)
		r.Delete("/unban", unban)
		r.Get("/poll", poll)
	})

	r.Post("/login", login)

	r.Mount("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	loadConfig()

	infoLog.Println("Fail2Ban jail set to " + conf.Jail)
	infoLog.Println("Listening port set to " + conf.Port)

	fmt.Println("Server started..\nListening on http://127.0.0.1:" + conf.Port)
	errorLog.Fatalln(http.ListenAndServe(":"+conf.Port, context.ClearHandler(r)))
}
