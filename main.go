package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"html/template"
	"log"
	"net/http"
)

const (
	TmplDir   = "templates"
	PublicDir = "public"
)

var (
	db *sql.DB
)

func init() {
	_db, err := sql.Open("postgres", "host=localhost dbname=DBNAME user=USERNAME password=PASSWORD sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	err = _db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	db = _db
}

func index(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(TmplDir + "/index.gohtml"))
	tmpl.Execute(w, nil)
}

func query(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	q := r.PostFormValue("sql")

	rows, err := db.Query(q)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))

	for i := range values {
		scanArgs[i] = &values[i]
	}

	data := struct {
		Columns []string
		Values  [][]string
	}{Columns: columns}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}
		var row []string
		var value string
		for _, val := range values {
			if val != nil {
				value = string(val)
			} else {
				value = "NULL"
			}
			row = append(row, value)
		}
		data.Values = append(data.Values, row)
	}

	tmpl := template.Must(template.ParseFiles(TmplDir + "/result.gohtml"))
	tmpl.Execute(w, data)
}

func main() {
	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(PublicDir))))
	r.HandleFunc("/", index)
	r.HandleFunc("/query", query)

	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: r,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
