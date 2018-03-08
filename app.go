package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
	"training/big_project/visitor"

	_ "github.com/lib/pq"
	nsq "github.com/nsqio/go-nsq"
)

type (
	// User for db
	User struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		MSISDN     int    `json:"msisdn"`
		Email      string `json:"email"`
		Birthdate  string `json:"birthdate"`
		CreateTime string `json:"created_time"`
		UpdateTime string `json:"update_time"`
		UserAge    int    `json:"user_age"`
	}
)

var templates *template.Template
var dbHome *sql.DB
var err error
var stmt *sql.Stmt

func main() {
	dbHome, err = sql.Open("postgres", "postgres://st140804:apaajadeh@devel-postgre.tkpd/tokopedia-user?sslmode=disable")
	if err != nil {
		log.Print(err)
	}
	preparedStatement()
	visitor.InitRedis()

	users := getMultipleUser("")

	t := template.Must(template.ParseFiles("./templates/home.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		visitLog()
		visitor := visitor.GetVisitor()
		data := struct {
			Users   []User
			Visitor int
		}{
			Users:   users,
			Visitor: visitor,
		}
		t.Execute(w, data)
	})

	http.HandleFunc("/users", usersHandler)
	http.ListenAndServe(":8050", nil)

	dbHome.Close()
}

func generateDate(a time.Time) (formatDateTime string) {
	layout := "02 Jan 2006 15:04:05"
	formatDateTime = a.Format(layout)
	return
}

func preparedStatement() {
	var err error
	q := `SELECT user_id,full_name,msisdn,user_email,birth_date,create_time,update_time 
			 FROM ws_user WHERE lower(full_name) LIKE $1 ORDER BY full_name ASC LIMIT 500`

	stmt, err = dbHome.Prepare(q)

	if err != nil {
		log.Print(err)
		return
	}
}

func visitLog() {
	config := nsq.NewConfig()
	p, err := nsq.NewProducer("devel-go.tkpd:4150", config)
	if err != nil {
		log.Panic(err)
	}
	err = p.Publish("write_test", []byte("User visit the page."))
	if err != nil {
		log.Panic(err)
	}
}

func getMultipleUser(param string) (result []User) {

	rows, err := stmt.Query("%" + strings.ToLower(param) + "%")
	if err != nil {
		log.Print(err)
		return
	}
	defer rows.Close()

	userList := []User{}
	var updateTime sql.NullString
	var birthdate sql.NullString
	var createTime time.Time

	for rows.Next() {
		u := &User{}
		err := rows.Scan(&u.ID, &u.Name, &u.MSISDN, &u.Email, &birthdate, &createTime, &updateTime)
		if err != nil {
			log.Print(err)
		}

		u.CreateTime = generateDate(createTime)

		if !updateTime.Valid {
			u.UpdateTime = "-"
		} else {
			u.UpdateTime = updateTime.String
		}

		if !birthdate.Valid {
			u.Birthdate = "-"
		} else {
			birthyear, _ := time.Parse(time.RFC3339, birthdate.String)
			u.Birthdate = birthdate.String
			u.UserAge = time.Now().Year() - birthyear.Year()
		}

		userList = append(userList, *u)
	}

	return userList
}

func usersHandler(w http.ResponseWriter, r *http.Request) {

	keys, ok := r.URL.Query()["name"]
	if !ok || len(keys) < 1 {
		log.Println("Url Param 'key' is missing")
		return
	}

	data := getMultipleUser(keys[0])

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(data)
}
