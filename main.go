package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

const ENV_VAR_TODO_MODE = "GOTODO_MODE"
const ENV_VAR_DYNDB_TABLE = "GOTODO_DYNDB_TABLE"

func newTodo(mode string, table string) (*TodoListAPI, error) {
	switch mode {
	case "dyndb":
		log.Printf("Using AWS DynamoDB " + table)
		return NewDynDBTodoList(table)
	default:
		return NewRAMTodoList()
	}
}

func newApp() *fiber.App {
	tlist, err := newTodo(os.Getenv(ENV_VAR_TODO_MODE), os.Getenv(ENV_VAR_DYNDB_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	app := fiber.New()
	app.Get("/list/:id", tlist.GetList)
	app.Get("/list/:id/items", tlist.GetListItems)
	app.Get("/todos/:id", tlist.GetTodo)
	app.Post("/todos/:id/done", tlist.SetTodoDone)
	app.Post("/todos", tlist.AddTodo)
	app.Post("/lists", tlist.NewList)
	app.Get("/healthcheck", tlist.HealthCheck)
	return app
}

func main() {
	newApp().Listen(":8000")
}
