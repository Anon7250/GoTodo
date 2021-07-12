package todos

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TodoItem struct {
	Id     string `json:"id"`
	Title  string `json:"title"`
	ListId string `json:"list_id"`
}

type TodoList struct {
	Id        string `json:"id"`
	TodoChunk string `json:"todo_chunk"`
}

type TodoChunk struct {
	Todos []string `json:"todos"`
	Next  string   `json:"next"`
}

type KeyValueDB interface {
	HasKey(key string) (bool, error)
	SetJson(key string, value interface{}) error
	GetJson(key string, valueOut interface{}) error

	// Transaction only goes through if values in conditions didn't change
	TransactSetJsons(writes map[string]interface{}, conditions map[string]interface{}) error

	// TODO: This is very expensive for AWS. Get rid of it
	ListJsons(keyPrefix string, valuesOut interface{}) error
}

type TodoListAPI struct {
	db KeyValueDB
}

func (todo *TodoListAPI) NewList(c *fiber.Ctx) error {
	chunk_key, chunk_id, err := todo.newKeyAndId("/todo_chunk/")
	if err != nil {
		return err
	}

	list_key, list_id, err := todo.newKeyAndId("/list/")
	if err != nil {
		return err
	}

	chunk := TodoChunk{nil, ""}
	list := TodoList{list_id, chunk_id}

	err = todo.db.SetJson(chunk_key, chunk)
	if err != nil {
		return err
	}
	err = todo.db.SetJson(list_key, list)
	if err != nil {
		return err
	}

	fmt.Println("Adding List: ", list_key, list)
	return c.JSON(list)
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

	todoKey, todoId, err := todo.newKeyAndId("/todo/")
	if err != nil {
		return err
	}
	item.Id = todoId
	fmt.Println("Adding item: ", item)

	if item.ListId == "" {
		return fiber.NewError(fiber.StatusForbidden, "TodoItem.list_id must not be empty")
	}

	fmt.Println("Checking list", "/list/"+item.ListId)
	exists, err := todo.db.HasKey("/list/" + item.ListId)
	if err != nil {
		return err
	}
	if !exists {
		return fiber.NewError(fiber.StatusForbidden, "Non existent todo list: "+item.ListId)
	}

	err = todo.db.SetJson(todoKey, item)
	if err != nil {
		return err
	}

	var todoList TodoList
	err = todo.db.GetJson("/list/"+item.ListId, &todoList)
	if err != nil {
		return err
	}

	var todoChunk TodoChunk
	err = todo.db.GetJson("/todo_chunk/"+todoList.TodoChunk, &todoChunk)
	if err != nil {
		return err
	}
	todoChunk.Todos = append(todoChunk.Todos, item.Id)

	return todo.db.SetJson("/todo_chunk/"+todoList.TodoChunk, &todoChunk)
}

func (todo *TodoListAPI) newKeyAndId(keyPrefix string) (string, string, error) {
	var err error
	var newUUID string
	for {
		newUUID, err = GetUUID()
		if err != nil {
			return "", "", err
		}
		key := keyPrefix + newUUID
		duplicate, err := todo.db.HasKey(key)
		if err != nil {
			return "", "", err
		}
		if !duplicate {
			break
		}
	}
	return keyPrefix + newUUID, newUUID, nil
}

var GetUUID = GetUUIDImpl

func GetUUIDImpl() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return id.URN(), nil
}
