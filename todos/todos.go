package todos

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TodoItem struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}

type KeyValueDB interface {
	HasKey(key string) (bool, error)
	SetJson(key string, value interface{}) error
	GetJson(key string, valueOut interface{}) error

	// TODO: This is very expensive for AWS. Get rid of it
	ListJsons(keyPrefix string, valuesOut interface{}) error
}

type TodoListAPI struct {
	db KeyValueDB
}

func (todo *TodoListAPI) GetAll(c *fiber.Ctx) error {
	var todoItems []TodoItem
	err := todo.db.ListJsons("/todo/", &todoItems)
	if err != nil {
		return err
	}

	todoIds := make([]string, 0)
	for _, todoItem := range todoItems {
		todoIds = append(todoIds, todoItem.Id)
	}
	return c.JSON(todoIds)
}

func (todo *TodoListAPI) GetById(c *fiber.Ctx) error {
	id := c.Params("id")
	var todoItem TodoItem
	err := todo.db.GetJson("/todo/"+id, &todoItem)
	if err != nil {
		return err
	}
	return c.JSON(todoItem)
}

func (todo *TodoListAPI) AddTodo(c *fiber.Ctx) error {
	fmt.Println("Parsing item...")

	item := new(TodoItem)
	if err := c.BodyParser(item); err != nil {
		return err
	}

	newId, err := todo.newKey("")
	if err != nil {
		return err
	}
	item.Id = newId
	fmt.Println("Adding item: ", item)
	return todo.db.SetJson("/todo/"+item.Id, item)
}

func (todo *TodoListAPI) newKey(keyPrefix string) (string, error) {
	var key string
	for {
		newUUID, err := GetUUID()
		if err != nil {
			return "", err
		}
		key = keyPrefix + newUUID
		duplicate, err := todo.db.HasKey(key)
		if err != nil {
			return "", err
		}
		if !duplicate {
			break
		}
	}
	return key, nil
}

var GetUUID = GetUUIDImpl

func GetUUIDImpl() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return id.URN(), nil
}
