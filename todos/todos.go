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
	Name      string `json:"name"`
	TodoChunk string `json:"todo_chunk,omitempty"`
}

type TodoChunk struct {
	Todos []string `json:"todos"`
	Next  string   `json:"next"`
}

type WriteTransaction struct {
	// Create json items that must not already exist
	creates map[string]interface{}

	// Append strings to lists of strings
	strListAppends map[string][]string

	// Create empty lists of strings
	strListCreates []string
}

type KeyValueDB interface {
	HasKey(key string) (bool, error)
	GetJson(key string, valueOut interface{}) error
	GetStringList(key string, valueOut *[]string) error
	DoWriteTransaction(transaction WriteTransaction) error
}

type TodoListAPI struct {
	db KeyValueDB
}

func (todo *TodoListAPI) NewList(c *fiber.Ctx) error {
	fmt.Println("Parsing new list...")

	inputList := new(TodoList)
	if err := c.BodyParser(inputList); err != nil {
		return err
	}

	chunk_key, chunk_id, err := todo.newKeyAndId("/todo_chunk/")
	if err != nil {
		return err
	}

	list_key, list_id, err := todo.newKeyAndId("/list/")
	if err != nil {
		return err
	}

	chunk := TodoChunk{nil, ""}
	list := TodoList{list_id, inputList.Name, chunk_id}

	err = todo.db.DoWriteTransaction(
		WriteTransaction{
			creates: map[string]interface{}{
				chunk_key: chunk,
				list_key:  list,
			},
			strListCreates: []string{chunk_key, list_key},
		},
	)
	if err != nil {
		return err
	}
	return todo.respondWithList(c, list)
}

func (todo *TodoListAPI) GetList(c *fiber.Ctx) error {
	id := c.Params("id")
	var todoList TodoList
	err := todo.db.GetJson("/list/"+id, &todoList)
	if err != nil {
		return err
	}
	return todo.respondWithList(c, todoList)
}

func (todo *TodoListAPI) GetListItems(c *fiber.Ctx) error {
	id := c.Params("id")
	var todoList TodoList
	var todoListItems []string
	var todoChunkItems []string
	err := todo.db.GetJson("/list/"+id, &todoList)
	if err != nil {
		return err
	}

	err = todo.db.GetStringList("/list/"+id, &todoListItems)
	if err != nil {
		return err
	}

	// TODO: return more than just the first chunk
	err = todo.db.GetStringList("/todo_chunk/"+todoList.TodoChunk, &todoChunkItems)
	if err != nil {
		return err
	}
	return c.JSON(todoChunkItems)
}

func (todo *TodoListAPI) GetTodo(c *fiber.Ctx) error {
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

	var todoList TodoList
	err = todo.db.GetJson("/list/"+item.ListId, &todoList)
	if err != nil {
		return err
	}

	// TODO: Don't always add to the first chunk in the list
	chunkKey := "/todo_chunk/" + todoList.TodoChunk
	return todo.db.DoWriteTransaction(
		WriteTransaction{
			creates: map[string]interface{}{
				todoKey: item,
			},
			strListAppends: map[string][]string{
				chunkKey: {item.Id},
			},
		},
	)
}

func (todo *TodoListAPI) HealthCheck(c *fiber.Ctx) error {
	return c.JSON("ok")
}

func (todo *TodoListAPI) newKeyAndId(keyPrefix string) (string, string, error) {
	newUUID, err := GetUUID()
	if err != nil {
		return "", "", err
	}
	return keyPrefix + newUUID, newUUID, nil
}

func (todo *TodoListAPI) respondWithList(c *fiber.Ctx, list TodoList) error {
	list.TodoChunk = "" // Do not expose internal implementation details
	return c.JSON(list)
}

var GetUUID = GetUUIDImpl

func GetUUIDImpl() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return id.URN(), nil
}
