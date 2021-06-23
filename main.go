package main

import (
	"log"
	"os"

	"github.com/Anon7250/gotodo/todos"
	"github.com/gofiber/fiber/v2"
)

const ENV_VAR_TODO_MODE = "GOTODO_MODE"

func newTodo(mode string) (*todos.TodoList, error) {
	switch mode {
	case "dyndb":
		log.Printf("Using AWS DynamoDB")
		return todos.NewDynDBTodoList()
	default:
		return todos.NewRAMTodoList()
	}
}

func newApp() *fiber.App {
	tlist, err := newTodo(os.Getenv(ENV_VAR_TODO_MODE))
	if err != nil {
		log.Fatal(err)
	}
	app := fiber.New()
	app.Get("/todos", tlist.GetAll)
	app.Get("/todos/:id", tlist.GetById)
	app.Post("/todos", tlist.AddTodo)
	return app
}

func main() {
	newApp().Listen(":8000")
}
