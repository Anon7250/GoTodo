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

func TestGetTodosFromInvalidList(t *testing.T) {
	NewTest(newApp()).
		Get("/list/bad_id").
		Expect(t).
		Status(http.StatusNotFound).
		End()
	NewTest(newApp()).
		Get("/list/bad_id/items").
		Expect(t).
		Status(http.StatusNotFound).
		End()
}

func TestAddTodoWithoutListForbidden(t *testing.T) {
	NewTest(newApp()).
		Post("/todos").
		Header("Content-Type", "application/json").
		Body(`{"title": "something", "list_id": "BAD_ID"}`).
		Expect(t).
		Status(http.StatusForbidden).
		End()
	NewTest(newApp()).
		Post("/todos").
		Header("Content-Type", "application/json").
		Body(`{"title": "something"}`).
		Expect(t).
		Status(http.StatusForbidden).
		End()
}

func TestAddTodoOk(t *testing.T) {
	app := newApp()

	var list todos.TodoList
	NewTest(app).
		Post("/lists").
		Header("Content-Type", "application/json").
		Body(`{}`).
		Expect(t).
		Status(http.StatusOK).
		End().JSON(&list)

	NewTest(app).
		Post("/todos").
		Header("Content-Type", "application/json").
		Body(`{"title": "something", "list_id": "` + list.Id + `"}`).
		Expect(t).
		Body("").
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
		Post("/lists").
		Header("Content-Type", "application/json").
		Body(`{"name": "test123"}`).
		Expect(t).
		Body(`{"id": "fakeid1", "name": "test123"}`).
		Status(http.StatusOK).
		End()

	NewTest(app).
		Post("/todos").
		Header("Content-Type", "application/json").
		Body(`{"title": "something", "list_id": "fakeid1"}`).
		Expect(t).
		Status(http.StatusOK).
		End()
	NewTest(app).
		Get("/list/fakeid1").
		Expect(t).
		Body(`{"id": "fakeid1", "name": "test123"}`).
		Status(http.StatusOK).
		End()
	NewTest(app).
		Get("/list/fakeid1/items").
		Expect(t).
		Body(`["fakeid1"]`).
		Status(http.StatusOK).
		End()
	NewTest(app).
		Get("/todos/fakeid1").
		Expect(t).
		Body(`{"id": "fakeid1", "title": "something", "list_id": "fakeid1"}`).
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
