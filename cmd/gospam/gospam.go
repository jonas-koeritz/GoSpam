package main

import (
	"context"
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/emersion/go-smtp"
	"github.com/jonas-koeritz/gospam"
	"github.com/spf13/viper"
	"github.com/tjarratt/babble"
)

type contextKey string

//go:embed static
var staticFS embed.FS

//go:embed templates
var templatesFS embed.FS

func main() {
	err := readInConfig()
	if err != nil {
		log.Printf("no config file, using defaults (%s)\n", err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	servicesWaitGroup := &sync.WaitGroup{}
	ctx, shutdown := context.WithCancel(context.Background())
	serviceContext := context.WithValue(ctx, contextKey("wg"), servicesWaitGroup)

	// Create an InMemoryBackend to store messages
	backend := &gospam.InMemoryBackend{
		MaxStoredMessage: viper.GetInt("MaxStoredMessages"),
	}

	servicesWaitGroup.Add(3)
	go smtpServer(serviceContext, backend)
	go mailboxCleanup(serviceContext, backend)
	go webServer(serviceContext, backend)

	<-sigs
	log.Printf("received signal, shutting down\n")
	shutdown()
	log.Printf("waiting for services\n")
	servicesWaitGroup.Wait()
}

func smtpServer(ctx context.Context, backend gospam.Backend) {
	defer ctx.Value(contextKey("wg")).(*sync.WaitGroup).Done()
	// Create and configure the SMTP listener
	s := smtp.NewServer(backend)
	s.Addr = viper.GetString("SMTPListenAddress")
	s.Domain = viper.GetString("Domain")
	s.ReadTimeout = time.Duration(viper.GetInt("SMTPTimeout")) * time.Second
	s.WriteTimeout = time.Duration(viper.GetInt("SMTPTimeout")) * time.Second
	s.MaxMessageBytes = viper.GetInt("MaximumMessageSize")
	s.MaxRecipients = viper.GetInt("MaxRecipients")

	// no authentication required to deliver email
	s.AuthDisabled = true
	s.AllowInsecureAuth = false

	log.Printf("starting SMTP server at %s\n", s.Addr)
	go s.ListenAndServe()

	<-ctx.Done()
	log.Printf("shutting down SMTP server\n")
	s.Close()
}

func mailboxCleanup(ctx context.Context, backend *gospam.InMemoryBackend) {
	defer ctx.Value(contextKey("wg")).(*sync.WaitGroup).Done()

	cleanupInterval := time.NewTicker(time.Duration(viper.GetInt("CleanupPeriod")) * time.Minute)
	retentionHours := viper.GetInt("RetentionHours")
	log.Printf("starting periodic cleanup task\n")
	for {
		select {
		case <-cleanupInterval.C:
			deadline := time.Now().Add(time.Duration(-retentionHours) * time.Hour)
			log.Printf("deleting all messages received before %s\n", deadline)
			backend.Cleanup(deadline)
		case <-ctx.Done():
			log.Printf("shutting down cleanup task")
			return
		}
	}
}

func webServer(ctx context.Context, backend *gospam.InMemoryBackend) {
	defer ctx.Value(contextKey("wg")).(*sync.WaitGroup).Done()

	staticFiles := http.FileServer(http.FS(staticFS))

	http.HandleFunc("/", indexView())
	http.HandleFunc("/mailbox", mailboxView(backend))
	http.HandleFunc("/mail", emlDownload(backend))
	http.Handle("/static/", staticFiles)

	httpListenAddress := viper.GetString("HTTPListenAddress")

	log.Printf("starting web interface at %s\n", httpListenAddress)
	httpServer := &http.Server{Addr: httpListenAddress}

	go httpServer.ListenAndServe()
	<-ctx.Done()
	log.Printf("shutting down web interface\n")
	httpServer.Shutdown(context.Background())
}

func indexView() func(http.ResponseWriter, *http.Request) {
	indexTemplate := template.Must(template.ParseFS(templatesFS, "templates/index.html"))

	domain := viper.GetString("Domain")
	aliasGenerator := aliasPlaceholderGenerator()

	return func(w http.ResponseWriter, r *http.Request) {
		indexTemplate.Execute(w, struct {
			Domain      string
			RandomAlias string
		}{
			Domain:      domain,
			RandomAlias: aliasGenerator(),
		})
	}
}

func aliasPlaceholderGenerator() func() string {
	if viper.GetBool("RandomAliasPlaceholder") {
		babbler := babble.NewBabbler()
		babbler.Count = 1
		aliasChars := regexp.MustCompile("[^a-zA-Z0-9]")
		return func() string {
			return strings.ToLower(aliasChars.ReplaceAllString(babbler.Babble(), ""))
		}
	} else {
		return func() string {
			return "alias"
		}
	}
}

func emlDownload(backend *gospam.InMemoryBackend) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		e := backend.GetEmailById(id)
		if e == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "message/rfc822")
		w.Header().Set("Content-Disposition", "attachment; filename="+e.From+".eml")
		w.Write(e.Data)
	}
}

func mailboxView(backend *gospam.InMemoryBackend) func(w http.ResponseWriter, r *http.Request) {
	mailboxTemplate, err := template.New("mailbox.html").Funcs(template.FuncMap{
		"DateFormat": func(date time.Time) string {
			return date.Format(time.RFC3339)
		},
		"Join": func(elements []string) string {
			return strings.Join(elements, ", ")
		},
		"ShowMail": showMail,
	}).ParseFS(templatesFS, "templates/mailbox.html")

	if err != nil {
		log.Printf("Error parsing template: %s\n", err)
	}

	aliasGenerator := aliasPlaceholderGenerator()

	retentionHours := viper.GetInt("RetentionHours")
	domain := viper.GetString("Domain")

	return func(w http.ResponseWriter, r *http.Request) {
		alias := r.URL.Query().Get("alias")
		if len(alias) == 0 {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		err := mailboxTemplate.Execute(w, struct {
			Alias          string
			RandomAlias    string
			Domain         string
			RetentionHours int
			EMails         []*gospam.EMail
		}{
			Alias:          alias,
			RandomAlias:    aliasGenerator(),
			Domain:         domain,
			RetentionHours: retentionHours,
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

func readInConfig() error {
	viper.SetConfigName("gospam.conf")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/gospam")
	viper.AddConfigPath("$HOME/.gospam")
	viper.AddConfigPath(".")

	viper.SetDefault("SMTPListenAddress", ":25")
	viper.SetDefault("HTTPListenAddress", ":80")
	viper.SetDefault("Domain", "localhost")
	viper.SetDefault("MaxStoredMessages", 100000)
	viper.SetDefault("CleanupPeriod", 5)
	viper.SetDefault("RetentionHours", 4)
	viper.SetDefault("MaximumMessageSize", 5*1024*1024)
	viper.SetDefault("SMTPTimeout", 60)
	viper.SetDefault("MaxRecipients", 10)
	viper.SetDefault("RandomAliasPlaceholder", false)

	return viper.ReadInConfig()
}
