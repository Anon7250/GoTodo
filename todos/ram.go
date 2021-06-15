package todos

import (
	"github.com/gofiber/fiber/v2"
)

type RAMTodoList map[string]TodoItem

func NewRAMTodoList() (*TodoList, error) {
	ramlist := make(RAMTodoList, 0)
  return &TodoList {impl: &ramlist}, nil
}

func (todo *RAMTodoList) GetAll() ([]string, error) {
  ids := make([]string, 0)
  for id, _ := range *todo {
    ids = append(ids, id)
  }
	return ids, nil
}

func (todo *RAMTodoList) GetById(id string) (*TodoItem, error) {
  if item, found := (*todo)[id]; found {
    return &item, nil
  }
  return nil, fiber.NewError(fiber.StatusNotFound)
}

func (todo *RAMTodoList) AddTodo(item TodoItem) error {
  id, err := GetUUID()
  if err != nil {
    return err
  }

  item.Id = id
  (*todo)[item.Id] = item
	return nil
}
