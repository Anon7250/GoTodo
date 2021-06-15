// This is an integration test
package main

import (
	"io"
	"net/http"
	"testing"

	"github.com/Anon7250/gotodo/todos"
	"github.com/gofiber/fiber/v2"
  "github.com/prashantv/gostub"
	"github.com/steinfletcher/apitest"
)


func TestGetAllTodosInitialValue(t *testing.T) {
  NewTest(newApp()).
    Get("/todos").
    Expect(t).
    Body(`[]`).
    Status(http.StatusOK).
    End()
}

func TestAddTodoOk(t *testing.T) {
  NewTest(newApp()).
    Post("/todos").
    Header("Content-Type", "application/json").
    Body(`{"title": "something"}`).
    Expect(t).
    Status(http.StatusOK).
    End()
}

func TestAddTodoAndGetAll(t *testing.T) {
  app := newApp()

  stubs := gostub.Stub(&todos.GetUUID, func() (string, error) {
    return "fakeid1", nil
  })
  defer stubs.Reset()

  NewTest(app).
    Post("/todos").
    Header("Content-Type", "application/json").
    Body(`{"title": "something"}`).
    Expect(t).
    Status(http.StatusOK).
    End()
  NewTest(app).
    Get("/todos").
    Expect(t).
    Body(`["fakeid1"]`).
    Status(http.StatusOK).
    End()
  NewTest(app).
    Get("/todos/fakeid1").
    Expect(t).
    Body(`{"id": "fakeid1", "title": "something"}`).
    Status(http.StatusOK).
    End()
}

func NewTest(app *fiber.App) *apitest.APITest {
  return apitest.New().
                 Handler(FiberToHandlerFunc(app))
}

func FiberToHandlerFunc(app *fiber.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := app.Test(r)
		if err != nil {
			panic(err)
		}

		// copy headers
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)

		if _, err := io.Copy(w, resp.Body); err != nil {
			panic(err)
		}
	}
}
