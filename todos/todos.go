package todos
import (
  "fmt"

	"github.com/gofiber/fiber/v2"
)

type TodoItem struct {
  Id string `json:"id"`
	Title string `json:"title"`
}

type ITodoList interface {
  AddTodo(item TodoItem) error
  GetAll() ([]string, error)
  GetById(id string) (*TodoItem, error)
}

type TodoList struct {
  impl ITodoList
}

func (todo *TodoList) GetAll(c *fiber.Ctx) error {
  ans, err := todo.impl.GetAll()
  if err != nil {
    return err
  }
  return c.JSON(ans)
}

func (todo *TodoList) GetById(c *fiber.Ctx) error {
  id := c.Params("id")
  ans, err := todo.impl.GetById(id)
  if err != nil {
    return err
  }
  return c.JSON(ans)
}

func (todo *TodoList) AddTodo(c *fiber.Ctx) error {
	fmt.Println("Parsing item...")

	item := new(TodoItem)
	if err := c.BodyParser(item); err != nil {
		return err
	}

	fmt.Println("Adding item: ", item)
  return todo.impl.AddTodo(*item)
}
