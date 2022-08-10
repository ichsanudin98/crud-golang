package main

//import package and library go
import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"html/template"
	"log"
	"net/http"
	"time"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = ""         // password of database
	dbname   = "app_task" // name of database
)

var db *sql.DB

var err error

var tpl *template.Template

// conect db and set template
func init() {
	psqlConn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err = sql.Open("postgres", psqlConn)
	checkErr(err)
	err = db.Ping()
	checkErr(err)
	tpl = template.Must(template.ParseGlob("view/*"))
}

// function checked error
func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

// main function first execute
func main() {

	defer db.Close()
	// task
	http.HandleFunc("/", getTaskList)
	http.HandleFunc("/createTaskForm", createTaskForm)
	http.HandleFunc("/createTask", createTask)
	http.HandleFunc("/editTask", editTask)
	http.HandleFunc("/updateTask", updateTask)
	http.HandleFunc("/updateStatusTask", updateStatusTask)
	http.HandleFunc("/deleteTask", deleteTask)

	//run server in 127.0.0.1:9000
	log.Println("Server is up on 9000 port")
	log.Fatalln(http.ListenAndServe(":9000", nil))
}

// declaration table of task
type taskTable struct {
	ID       int64
	Task     string
	Assign   string
	Deadline string
	IsDone   bool
}

// struct for insert task
func (task taskTable) insertTask(taskDetail, assign, deadline string) (sql.Result, error) {
	task.Task = taskDetail
	task.Assign = assign
	task.Deadline = deadline

	queryInsert := `INSERT INTO task (task,assign,deadline) VALUES ($1,$2,$3)`
	return db.Exec(queryInsert,
		task.Task,
		task.Assign,
		task.Deadline,
	)
}

// index.html function
func getTaskList(w http.ResponseWriter, _ *http.Request) {
	rows, e := db.Query(
		`SELECT id, task, assign, deadline, is_done FROM task;`)

	if e != nil {
		log.Println(e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

	tasks := make([]taskTable, 0)
	for rows.Next() {
		data := taskTable{}
		rows.Scan(&data.ID, &data.Task, &data.Assign, &data.Deadline, &data.IsDone)
		layout := "2006-01-02T15:04:05Z0700"
		t, _ := time.Parse(layout, data.Deadline)
		data.Deadline = t.Format("02-01-2006")
		tasks = append(tasks, data)
	}
	tpl.ExecuteTemplate(w, "index.html", tasks)
}

// createTask.html function
func createTaskForm(w http.ResponseWriter, _ *http.Request) {
	err = tpl.ExecuteTemplate(w, "createTask.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// function to create task
func createTask(w http.ResponseWriter, req *http.Request) {
	t, _ := time.Parse("2006-01-02", req.FormValue("deadline"))
	d := t.Unix()
	x, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	today := x.Unix()

	if d < today {
		http.Error(w, "Waktu telah terlewati", http.StatusBadRequest)
		return
	}

	if req.Method == http.MethodPost {
		task := taskTable{}
		_, e := task.insertTask(
			req.FormValue("task"),
			req.FormValue("assign"),
			req.FormValue("deadline"))

		if e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
}

// editTask.html function
func editTask(w http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	rows, err := db.Query(
		`SELECT id, 
       task,
       assign,
       deadline,
       is_done FROM task WHERE id = ` + id + `;`)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := taskTable{}
	for rows.Next() {
		rows.Scan(&data.ID, &data.Task, &data.Assign, &data.Deadline, &data.IsDone)
	}
	layout := "2006-01-02T15:04:05Z0700"
	t, _ := time.Parse(layout, data.Deadline)
	data.Deadline = t.Format("2006-01-02")
	tpl.ExecuteTemplate(w, "editTask.html", data)
}

// function to update task
func updateTask(w http.ResponseWriter, req *http.Request) {
	queryUpdate := `UPDATE task SET task=$1, assign=$2, deadline=$3 WHERE id=$4`
	_, err := db.Exec(queryUpdate,
		req.FormValue("task"),
		req.FormValue("assign"),
		req.FormValue("deadline"),
		req.FormValue("id"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)
}

// function to update status task
func updateStatusTask(w http.ResponseWriter, req *http.Request) {
	queryUpdate := `UPDATE task SET is_done=$1 WHERE id=$2`
	_, err := db.Exec(queryUpdate,
		true,
		req.FormValue("id"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)
}

// function to delete task
func deleteTask(w http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	if id == "" {
		http.Error(w, "Please Send ID", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(`DELETE FROM task WHERE id=$1`, id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, req, "/", http.StatusSeeOther)
}
