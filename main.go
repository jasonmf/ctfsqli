package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func fatalIfError(err error, msg string) {
	if err != nil {
		log.Fatal("error ", msg, ": ", err)
	}
}

func main() {
	tFlag := os.Getenv("TARGET_FLAG")
	if tFlag == "" {
		log.Fatal("missing env var: TARGET_FLAG")
	}
	listen := ":8000"
	if v := os.Getenv("LISTEN"); v != "" {
		listen = v
	}
	templateDir := "out/templates"
	if v := os.Getenv("TEMPLATE_DIR"); v != "" {
		templateDir = v
	}

	templates, err := template.ParseGlob(templateDir + "/*.html")
	fatalIfError(err, "parsing templates")

	db, err := sql.Open("mysql", "ctfrw:dDIESeNBAARaMapY0kc3Q@unix(/var/run/mysqld/mysqld.sock)/ctf")
	fatalIfError(err, "opening database for writing")

	_, err = db.Exec("UPDATE users SET password = ? WHERE username = 'ellie'", tFlag)
	fatalIfError(err, "setting flag")
	fatalIfError(db.Close(), "closing DB")

	db, err = sql.Open("mysql", "ctfro:CB2fwpYY5c+KpT2FxzDmaA@unix(/var/run/mysqld/mysqld.sock)/ctf")
	fatalIfError(err, "reopening database")
	defer db.Close()

	server := Server{
		db:   db,
		tmpl: templates,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/expenses", server.AuthWrap(server.Expenses))
	mux.HandleFunc("/users", server.AuthWrap(server.Users))

	log.Print("listening on ", listen)
	fatalIfError(http.ListenAndServe(listen, mux), "listening")
}

type Server struct {
	db   *sql.DB
	tmpl *template.Template
}

// AuthWrap restricts calling the wrapped handler to authenticated requests
func (s Server) AuthWrap(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "joel" || pass != "n0Clikkerz" {
			w.Header().Set("WWW-Authenticate", "Basic realm=expenses")
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		f(w, r)
	}
}

func (s Server) Expenses(w http.ResponseWriter, r *http.Request) {
	what := r.FormValue("what")
	rows, err := s.db.Query("SELECT * FROM expenses WHERE what LIKE '%" + what + "%'")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	v := [][]interface{}{}
	for rows.Next() {
		cols, err := rows.Columns()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		rowVals := make([]interface{}, len(cols))
		for i := range rowVals {
			rowVals[i] = new(string)
		}
		if err := rows.Scan(rowVals...); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		v = append(v, rowVals)
	}
	w.Header().Set("Content-Type", "text/html")
	data := map[string]interface{}{
		"Rows": v,
	}
	if err := s.tmpl.ExecuteTemplate(w, "expenses.html", data); err != nil {
		log.Print("executing template: ", err)
	}
}

func (s Server) Users(w http.ResponseWriter, r *http.Request) {
	what := r.FormValue("who")
	rows, err := s.db.Query("SELECT username, '********', added FROM users WHERE username LIKE ?", "%"+what+"%")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	v := [][]interface{}{}
	for rows.Next() {
		cols, err := rows.Columns()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		rowVals := make([]interface{}, len(cols))
		for i := range rowVals {
			rowVals[i] = new(string)
		}
		if err := rows.Scan(rowVals...); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		v = append(v, rowVals)
	}
	w.Header().Set("Content-Type", "text/html")
	data := map[string]interface{}{
		"Rows": v,
	}
	if err := s.tmpl.ExecuteTemplate(w, "users.html", data); err != nil {
		log.Print("executing template: ", err)
	}
}

func sendJSON(w http.ResponseWriter, v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/json")
	w.Write(b)
}
