package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var (
	db             *sql.DB
	mailgunStorage chan receiveMsg
)

func fetchEmails() {
	for {
		msg := <-mailgunStorage

		subdomain := strings.Replace(msg.To, "@"+os.Getenv("MAILGUN_DOMAIN"), "", -1)

		db.Query("insert into posts (email, subdomain, subject, body, message_id) values ($1,$2,$3,$4,$5)", msg.From, subdomain, msg.Subject, msg.Body, msg.MessageID)
	}
}

type receiveMsg struct {
	To        string
	From      string
	Subject   string `json:"subject"`
	MessageID string `json:"Message-ID"`
	Body      string `json:"stripped-html"`
}

func receiveHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	log.Println(r.Form)

	msg := receiveMsg{
		To:        r.Form.Get("recipient"),
		From:      r.Form.Get("from"),
		Subject:   r.Form.Get("subject"),
		MessageID: r.Form.Get("Message-Id"),
		Body:      r.Form.Get("body-plain"),
	}

	mailgunStorage <- msg

	w.Write([]byte(""))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	hostSplit := strings.Split(r.Host, ".")
	subdomain := hostSplit[0]

	rows, err := db.Query("SELECT email, subject, body, created_at FROM posts WHERE subdomain = $1 limit 100", subdomain)
	if err != nil {
		w.Write([]byte("Nothing here govn'r"))
		return
	}

	type post struct {
		Email     string
		Subject   string
		Body      string
		CreatedAt *time.Time
	}

	var posts []post

	for rows.Next() {
		var p post
		rows.Scan(&p.Email, &p.Subject, &p.Body, &p.CreatedAt)

		posts = append(posts, p)
	}

	const tpl = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>{{.Title}}</title>
<link rel="stylesheet" href="https://bootswatch.com/paper/bootstrap.min.css"></link>
</head>
<body>
<div class="container">
<div class="row">
<div class="col-lg-push-3 col-lg-6">
<h1>{{.Title}}</h1>
<p>Send emails to {{.Title}}@journlr.com</p>
</div>
</div>
{{range .Posts}}
<div class="row">
<div class="col-lg-push-3 col-lg-6">
<h3>{{ .Subject }}</h3>
<pre>
{{ .Body }}
</pre>
{{ .CreatedAt }}
</div>
</div>
{{else}}
<div><strong>no posts</strong></div>{{end}}
</body>
</html>`

	data := struct {
		Title string
		Posts []post
	}{
		Title: subdomain,
		Posts: posts,
	}
	t, err := template.New("webpage").Parse(tpl)

	t.Execute(w, data)
}

func main() {
	log.Println("Starting journlr...")

	var err error
	db, err = sql.Open("postgres", os.Getenv("JOURNLR_DB"))

	if err != nil {
		log.Fatal("No database for journl")
	}

	r := mux.NewRouter()

	r.HandleFunc("/", indexHandler).Methods("GET")
	r.HandleFunc("/receive", receiveHandler).Methods("POST")

	mailgunStorage = make(chan receiveMsg)

	go fetchEmails()

	http.ListenAndServe(":6891", handlers.LoggingHandler(os.Stdout, r))
}
