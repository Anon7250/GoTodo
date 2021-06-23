package todos

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TodoItem struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}

type KeyValueDB interface {
	SetJson(key string, json []byte) error
	GetJson(key string) ([]byte, error)

	// TODO: This is very expensive for AWS. Get rid of it
	ListJsons(keyPrefix string) ([][]byte, error)
}

type TodoList struct {
	db KeyValueDB
}

func (todo *TodoList) GetAll(c *fiber.Ctx) error {
	ans, err := todo.db.ListJsons("/todo/")
	if err != nil {
		return err
	}

	var todoItems = make([]string, 0)
	for _, rawJson := range ans {
		var todoItem TodoItem
		err := json.Unmarshal(rawJson, &todoItem)
		if err != nil {
			return err
		}
		todoItems = append(todoItems, todoItem.Id)
	}
	return c.JSON(todoItems)
}

func (todo *TodoList) GetById(c *fiber.Ctx) error {
	id := c.Params("id")
	ans, err := todo.db.GetJson("/todo/" + id)
	if err != nil {
		return err
	}
	var todoItem TodoItem
	err = json.Unmarshal(ans, &todoItem)
	if err != nil {
		return err
	}
	return c.JSON(todoItem)
}

func (todo *TodoList) AddTodo(c *fiber.Ctx) error {
	fmt.Println("Parsing item...")

	item := new(TodoItem)
	if err := c.BodyParser(item); err != nil {
		return err
	}

	newUUID, err := GetUUID()
	if err != nil {
		return err
	}

	item.Id = newUUID
	fmt.Println("Adding item: ", item)

	jsonVal, err := json.Marshal(item)
	if err != nil {
		return err
	}
	return todo.db.SetJson("/todo/"+item.Id, jsonVal)
}

var GetUUID = GetUUIDImpl

func GetUUIDImpl() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return id.URN(), nil
}
