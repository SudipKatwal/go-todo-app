package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/thedevsaddam/renderer"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func helloWorldPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "todo.html")
}

var res *renderer.Render
var db *mgo.Database

const (
	hostname string = "localhost:27017"
	dbName   string = "db_todo"
	port     string = ":5000"
)

type (
	todoModel struct {
		ID        bson.ObjectId `bson:"_id,omitempty"`
		Title     string        `bson:"title"`
		CreatedAt time.Time     `bson:"createdAt"`
	}

	todo struct {
		ID        string    `json:"id"`
		Title     string    `json:"title"`
		CreatedAt time.Time `json:"created_at"`
	}
)

func init() {
	res = renderer.New()
	sess, err := mgo.Dial(hostname)
	checkErr(err)
	sess.SetMode(mgo.Monotonic, true)

	db = sess.DB(dbName)

}

func homePage(w http.ResponseWriter, r *http.Request) {
	err := res.Template(w, http.StatusOK, []string{"pages/todo.html"}, nil)
	checkErr(err)
}

func fetchTodos(w http.ResponseWriter, r *http.Request) {

}

func createTodo(w http.ResponseWriter, r *http.Request) {
	var t todo

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		res.JSON(w, http.StatusProcessing, err)
		return
	}

	// validation
	if t.Title == "" {
		res.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The title field is required",
		})
		return
	}

	// if input is okay, create a todo
	tm := todoModel{
		ID:        bson.NewObjectId(),
		Title:     t.Title,
		CreatedAt: time.Now(),
	}

	if err := db.C("todo").Insert(&tm); err != nil {
		res.JSON(w, http.StatusProcessing, renderer.M{
			"message": "Failed to save todo",
			"error":   err,
		})
		return
	}

	res.JSON(w, http.StatusCreated, renderer.M{
		"message": "Todo created successfully",
		"todo_id": tm.ID.Hex(),
	})
}

func updateTodo(w http.ResponseWriter, r *http.Request) {

}

func deleteTodo(w http.ResponseWriter, r *http.Request) {

}

func main() {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", homePage)

	r.Mount("/todo", routeHandlers())

	srv := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("Listening on port", port)
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("Listen : %s\n", err)
		}
	}()

	<-stopChan
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)
	defer cancel()
	log.Println("Server gracefully stopped")

}

func routeHandlers() http.Handler {
	rg := chi.NewRouter()
	rg.Group(func(r chi.Router) {
		r.Get("/", fetchTodos)
		r.Post("/", createTodo)
		r.Put("/{id}", updateTodo)
		r.Delete("/{id}", deleteTodo)
	})
	return rg
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err) //respond with error page or message
	}
}
