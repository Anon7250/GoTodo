package todos

import (
	"github.com/gofiber/fiber/v2"
  "github.com/google/uuid"
)

type RAMTodoList map[string]TodoItem

var GetUUID = GetUUIDImpl

func NewRAMTodoList() TodoList {
	ramlist := make(RAMTodoList, 0)
  return TodoList {impl: &ramlist}
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

func GetUUIDImpl() (string, error) {
  id, err := uuid.NewRandom()
  if err != nil {
    return "", err
  }
  return id.URN(), nil
}
