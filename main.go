package main

import (
	"github.com/Anon7250/gotodo/todos"
	"github.com/gofiber/fiber/v2"
)

func newApp() *fiber.App {
	tlist := todos.NewRAMTodoList()
	app := fiber.New()
	app.Get("/todos", tlist.GetAll)
	app.Get("/todos/:id", tlist.GetById)
	app.Post("/todos", tlist.AddTodo)
	return app
}

func main() {
	newApp().Listen(":8000")
}
