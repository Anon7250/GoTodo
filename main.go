package main

import (
	"github.com/Anon7250/gotodo/todos"
	"github.com/gofiber/fiber/v2"
)

func newApp() *fiber.App {
	tlist := make(todos.TodoList, 0)
	app := fiber.New()
	app.Get("/todos", tlist.GetAll)
	app.Post("/todos", tlist.AddTodo)
	return app
}

func main() {
	newApp().Listen("localhost:8000")
}
