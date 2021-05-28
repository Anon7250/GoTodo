package todos

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type TodoItem struct {
	Title string `json:"title"`
}

type TodoList []TodoItem

func (todo *TodoList) GetAll(c *fiber.Ctx) error {
	return c.JSON(todo)
}

func (todo *TodoList) AddTodo(c *fiber.Ctx) error {
	fmt.Println("Parsing item...")

	item := new(TodoItem)
	if err := c.BodyParser(item); err != nil {
		return err
	}

	fmt.Println("Adding item: ", item)
  *todo = append(*todo, *item)
	return nil
}
