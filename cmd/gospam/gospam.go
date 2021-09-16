package main

import (
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-smtp"
	"github.com/jonas-koeritz/gospam"
	"github.com/spf13/viper"
	"github.com/tjarratt/babble"
)

func main() {
	err := readInConfig()
	if err != nil {
		log.Printf("Error reading config: %s\n", err)
		return
	}
	backend := &gospam.Backend{
		MaxStoredMessage: viper.GetInt("MaxStoredMessages"),
	}
	s := smtp.NewServer(backend)

	s.Addr = viper.GetString("SMTPListenAddress")
	s.Domain = viper.GetString("Domain")
	s.ReadTimeout = 60 * time.Second
	s.WriteTimeout = 60 * time.Second
	s.MaxMessageBytes = 5 * 1024 * 1024 // 5 MiB
	s.MaxRecipients = 10
	s.AllowInsecureAuth = true

	go mailboxCleanup(backend)

	go func() {
		log.Printf("Starting Mailserver at %s\n", s.Addr)
		if err := s.ListenAndServe(); err != nil {
			log.Printf("Error: %s\n", err)
			return
		}
	}()
	defer s.Close()

	webServer(backend)
}

func webServer(backend *gospam.Backend) {
	mux := http.NewServeMux()
	staticFiles := http.FileServer(http.Dir("./static/"))

	mux.HandleFunc("/", indexView())
	mux.HandleFunc("/mailbox", mailboxView(backend))
	mux.HandleFunc("/mail", emlDownload(backend))
	mux.Handle("/static/", http.StripPrefix("/static", staticFiles))

	log.Printf("Web interface listening at %s\n", viper.GetString("HTTPListenAddress"))
	http.ListenAndServe(viper.GetString("HTTPListenAddress"), mux)
}

func indexView() func(http.ResponseWriter, *http.Request) {
	indexTemplate := template.Must(template.ParseFiles("./templates/index.html"))
	babbler := babble.NewBabbler()
	babbler.Count = 1

	return func(w http.ResponseWriter, r *http.Request) {

		indexTemplate.Execute(w, struct {
			Domain      string
			RandomAlias string
		}{
			Domain:      viper.GetString("Domain"),
			RandomAlias: makeAlias(babbler.Babble()),
		})
	}
}

func emlDownload(backend *gospam.Backend) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ids := r.URL.Query()["id"]
		id := ""
		if len(ids) > 0 {
			id = ids[0]
		}
		numericId, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			w.WriteHeader(404)
		}

		e := backend.GetEmailById(numericId)
		w.Header().Set("Content-Type", "message/rfc822")
		w.Header().Set("Content-Disposition", "attachment; filename="+e.From+".eml")
		w.Write(e.Data)
	}
}

func mailboxCleanup(backend *gospam.Backend) {
	for {
		time.Sleep(time.Duration(viper.GetInt("CleanupPeriod")) * time.Minute)
		backend.Cleanup(viper.GetInt("RetentionHours"))
	}
}

func mailboxView(backend *gospam.Backend) func(w http.ResponseWriter, r *http.Request) {
	mailboxTemplate, err := template.New("mailbox.html").Funcs(template.FuncMap{
		"DateFormat": func(date time.Time) string {
			return date.Format(time.RFC3339)
		},
		"Join": func(elements []string) string {
			return strings.Join(elements, ", ")
		},
		"Sanitize": func(data []byte) string {
			return template.HTMLEscapeString(string(data))
		},
		"ShowMail": showMail,
	}).ParseFiles("./templates/mailbox.html")

	if err != nil {
		log.Printf("Error parsing template: %s\n", err)
	}

	babbler := babble.NewBabbler()
	babbler.Count = 1

	return func(w http.ResponseWriter, r *http.Request) {
		aliases := r.URL.Query()["alias"]
		alias := ""
		if len(aliases) > 0 {
			alias = aliases[0]
		}

		err := mailboxTemplate.Execute(w, struct {
			Alias          string
			RandomAlias    string
			Domain         string
			RetentionHours int
			EMails         []*gospam.EMail
		}{
			Alias:          alias,
			RandomAlias:    makeAlias(babbler.Babble()),
			Domain:         viper.GetString("Domain"),
			RetentionHours: viper.GetInt("RetentionHours"),
			EMails:         backend.GetEmailsByAlias(alias),
		})
		if err != nil {
			log.Printf("ERROR: %s\n", err)
		}
	}
}

func showMail(email *gospam.EMail) string {
	return string(email.Data)
}

func makeAlias(candidate string) string {
	aliasChars := regexp.MustCompile("[^a-zA-Z0-9]")
	return strings.ToLower(aliasChars.ReplaceAllString(candidate, ""))
}

func readInConfig() error {
	viper.SetConfigName("gospam.conf")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/gospam")
	viper.AddConfigPath("$HOME/.gospam")
	viper.AddConfigPath(".")

	viper.SetDefault("SMTPListenAddress", ":25")
	viper.SetDefault("MaxStoredMessages", 100000)
	viper.SetDefault("CleanupPeriod", 5)
	viper.SetDefault("RetentionHours", 4)

	return viper.ReadInConfig()
}
