package main

import (
	"fmt"

	"github.com/Anon7250/gonorm"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const DefaultTableName = "GoTodo1"

func NewDynDBTodoList(table string) (*TodoListAPI, error) {
	if len(table) == 0 {
		table = DefaultTableName
	}
	db, err := gonorm.NewDynDB(table)
	if err != nil {
		return nil, err
	}
	return &TodoListAPI{db: db}, nil
}

func NewRAMTodoList() (*TodoListAPI, error) {
	return &TodoListAPI{db: gonorm.NewRAMDB()}, nil
}

type TodoItem struct {
	Done   bool   `json:"done" dynamodbav:"done"`
	Id     string `json:"id" dynamodbav:"id"`
	Title  string `json:"title" dynamodbav:"title"`
	ListId string `json:"list_id" dynamodbav:"list_id"`
}

type TodoList struct {
	Id        string `json:"id" dynamodbav:"id"`
	Name      string `json:"name" dynamodbav:"name"`
	TodoChunk string `json:"todo_chunk,omitempty" dynamodbav:"todo_chunk"`
}

type TodoChunk struct {
	Todos []string `json:"todos" dynamodbav:"todos"`
	Next  string   `json:"next" dynamodbav:"next"`
}

type TodoListAPI struct {
	db gonorm.KeyValueDB
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
		gonorm.WriteTransaction{
			Creates: map[string]interface{}{
				chunk_key: chunk,
				list_key:  list,
			},
			StrListCreates: []string{chunk_key, list_key},
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

type GetListItemsQuery struct {
	Done     *bool `query:"done"`
	Position *int  `query:"pos"`
	Length   *int  `query:"len"`
}

func (todo *TodoListAPI) GetListItems(c *fiber.Ctx) error {
	id := c.Params("id")
	var query GetListItemsQuery

	err := c.QueryParser(&query)
	if err != nil {
		return err
	}

	var todoList TodoList
	var todoListItems []string
	todoChunkItems := make([]string, 0)
	err = todo.db.GetJson("/list/"+id, &todoList)
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

	if query.Position != nil {
		if *query.Position >= len(todoChunkItems) {
			todoChunkItems = make([]string, 0)
		} else {
			todoChunkItems = todoChunkItems[*query.Position:]
		}
	}
	if query.Length != nil {
		if *query.Length < len(todoChunkItems) {
			todoChunkItems = todoChunkItems[:*query.Length]
		}
	}
	if query.Done != nil {
		keys := make([]string, 0)
		oldIds := todoChunkItems
		for _, id := range oldIds {
			keys = append(keys, "/todo/"+id)
		}

		valiIds := make(map[string]bool)
		var rawTodoItems []interface{}
		err := todo.db.GetJsons(keys, &rawTodoItems)
		if err != nil {
			return err
		}
		for _, rawTodoItem := range rawTodoItems {
			var todoItem TodoItem
			err := todo.db.Unmarshal(rawTodoItem, &todoItem)
			if err != nil {
				return err
			}
			if todoItem.Done != *query.Done {
				continue
			}
			valiIds[todoItem.Id] = true
		}

		todoChunkItems = make([]string, 0)
		for _, id := range oldIds {
			_, ok := valiIds[id]
			if ok {
				todoChunkItems = append(todoChunkItems, id)
			}
		}
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

func (todo *TodoListAPI) SetTodoDone(c *fiber.Ctx) error {
	id := c.Params("id")
	var done bool

	err := c.BodyParser(&done)
	if err != nil {
		return err
	}

	err = todo.db.DoWriteTransaction(gonorm.WriteTransaction{
		SetFields: map[string]map[string]interface{}{
			"/todo/" + id: {"done": done},
		},
	})
	if err != nil {
		return err
	}
	return c.JSON(map[string]string{})
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
	err = todo.db.DoWriteTransaction(
		gonorm.WriteTransaction{
			Creates: map[string]interface{}{
				todoKey: item,
			},
			StrListAppends: map[string][]string{
				chunkKey: {item.Id},
			},
		},
	)
	if err != nil {
		return err
	}
	return c.JSON(item)
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
