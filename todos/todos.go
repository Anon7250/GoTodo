package todos

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
  "github.com/google/uuid"
)

type TodoItem struct {
  Id string `json:"id"`
	Title string `json:"title"`
}

type TodoList map[string]TodoItem

var GetUUID = GetUUIDImpl

func (todo *TodoList) GetAll(c *fiber.Ctx) error {
  ids := make([]string, 0)
  for id, _ := range *todo {
    ids = append(ids, id)
  }
	return c.JSON(ids)
}

func (todo *TodoList) GetById(c *fiber.Ctx) error {
  id := c.Params("id")
  if item, found := (*todo)[id]; found {
    return c.JSON(item)
  }
  return fiber.NewError(fiber.StatusNotFound)
}

func (todo *TodoList) AddTodo(c *fiber.Ctx) error {
	fmt.Println("Parsing item...")

	item := new(TodoItem)
	if err := c.BodyParser(item); err != nil {
		return err
	}

	fmt.Println("Adding item: ", item)

  id, err := GetUUID()
  if err != nil {
    return err
  }

  item.Id = id
  (*todo)[item.Id] = *item
	return nil
}

func GetUUIDImpl() (string, error) {
  id, err := uuid.NewRandom()
  if err != nil {
    return "", err
  }
  return id.URN(), nil
}
