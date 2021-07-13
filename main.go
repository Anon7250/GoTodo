package main

import (
	"log"
	"os"

	"github.com/Anon7250/gotodo/todos"
	"github.com/gofiber/fiber/v2"
)

const ENV_VAR_TODO_MODE = "GOTODO_MODE"
const ENV_VAR_DYNDB_TABLE = "GOTODO_DYNDB_TABLE"

func newTodo(mode string, table string) (*todos.TodoListAPI, error) {
	switch mode {
	case "dyndb":
		log.Printf("Using AWS DynamoDB " + table)
		return todos.NewDynDBTodoList(table)
	default:
		return todos.NewRAMTodoList()
	}
}

func newApp() *fiber.App {
	tlist, err := newTodo(os.Getenv(ENV_VAR_TODO_MODE), os.Getenv(ENV_VAR_DYNDB_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	app := fiber.New()
	app.Get("/list/:id", tlist.GetListItems)
	app.Get("/todos/:id", tlist.GetTodo)
	app.Post("/todos", tlist.AddTodo)
	app.Post("/lists", tlist.NewList)
	app.Get("/healthcheck", tlist.HealthCheck)
	return app
}

func main() {
	newApp().Listen(":8000")
}
