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

var (
	conf     = new(config)
	infoLog  *log.Logger
	errorLog *log.Logger
	logF     *os.File
	store    = sessions.NewCookieStore([]byte(conf.LDAPKey))
)

var (
	tableTemplate  = template.Must(template.ParseFiles("static/main.html", "static/table.html"))
	formTemplate   = template.Must(template.ParseFiles("static/main.html", "static/form.html"))
	noauthTemplate = template.Must(template.ParseFiles("static/main.html", "static/noauth.html"))
)

func init() {
	store.Options = &sessions.Options{
		Domain:   conf.CookieHost,
		MaxAge:   60 * 10,
		HttpOnly: true,
	}

	gob.Register(user{})
}

func list(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data := struct {
		Data TableData
		IP   string
	}{
		renderTable(),
		r.RemoteAddr,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tableTemplate.ExecuteTemplate(w, "main", data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorLog.Printf("Failed to execute table template: %v", err)
	}
}

func unban(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
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
		errorLog.Printf("Failed to execute unban command: %s", b.String())
		return
	}

	infoLog.Printf("IP Address %s has been unbanned by %s", r.URL.Query()["ip"][0], u.Username)
}

func poll(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	lastDate := strings.TrimSuffix(strings.TrimPrefix(r.URL.Query()["date"][0], "\""), "\"")

	f, err := os.Open("/var/log/fail2ban.log")
	if err != nil {
		fmt.Fprint(w)
		errorLog.Printf("Failed to open fail2ban log: %v", err)
		return
	}
	defer f.Close()

	newLogText, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Fprint(w, "")
		errorLog.Printf("Failed to read fail2ban log: %v", err)
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
	defer r.Body.Close()
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
	defer r.Body.Close()
	data := struct {
		Data string
	}{
		msg,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := formTemplate.ExecuteTemplate(w, "main", data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorLog.Printf("Failed to execute login page template: %v", err)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	infoLog.Printf("login request from %s for username %s", r.RemoteAddr, username)

	user, err := getUserFromLDAP(username, password)
	if err != nil {
		switch {
		case err == errWrongPass || err == errNoUser:
			errorLog.Printf("IP %s failed login with username %s", r.RemoteAddr, username)
			http.Redirect(w, r, "/?auth=njet", http.StatusTemporaryRedirect)
			return
		default:
			errorLog.Printf("Failed to get user form LDAP: %v", err)
			http.Redirect(w, r, "/?auth=err", http.StatusTemporaryRedirect)
			return
		}
	}

	if !user.isadmin {
		http.Redirect(w, r, "/noauth", http.StatusTemporaryRedirect)
		infoLog.Printf("non-admin %s attempted login from %s", username, r.RemoteAddr)
		return
	}

	session, err := store.New(r, "id")
	if err != nil {
		http.Redirect(w, r, "/?auth=err", http.StatusTemporaryRedirect)
		errorLog.Printf("Failed to create new session: %v", err)
		return
	}

	session.Values["user"] = user

	if err := session.Save(r, w); err != nil {
		http.Redirect(w, r, "/?auth=err", http.StatusTemporaryRedirect)
		errorLog.Printf("error saving session: %v", err)
		return
	}

	infoLog.Printf("%s successfully logged in from %s", username, r.RemoteAddr)

	http.Redirect(w, r, "/list", http.StatusTemporaryRedirect)
}

func notAuthorized(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "text/html")
	if err := noauthTemplate.ExecuteTemplate(w, "main", nil); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorLog.Printf("Failed to execute not-authorized template: %v", err)
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

	fmt.Println(fmt.Sprintf("Server started..\nListening on http://%s:%s", conf.ListenHost, conf.Port))
	errorLog.Fatalln(http.ListenAndServe(":"+conf.Port, context.ClearHandler(r)))
}
