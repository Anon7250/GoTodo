// This is an integration test
package main

import (
	"fmt"
	"io"
	"net/http"
	"testing"

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

	var list TodoList
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

func TestOneTodoListWithOneItem(t *testing.T) {
	app := newApp()

	stubs := gostub.Stub(&GetUUID, func() (string, error) {
		return "fakeid1", nil
	})
	defer stubs.Reset()

	// Create a todo list
	NewTest(app).
		Post("/lists").
		Header("Content-Type", "application/json").
		Body(`{"name": "test123"}`).
		Expect(t).
		Body(`{"id": "fakeid1", "name": "test123"}`).
		Status(http.StatusOK).
		End()

	// Create a todo item
	NewTest(app).
		Post("/todos").
		Header("Content-Type", "application/json").
		Body(`{"title": "something", "list_id": "fakeid1"}`).
		Expect(t).
		Status(http.StatusOK).
		End()

	// Test all features of the todo list and the todo item
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
		Body(`{"done": false, "id": "fakeid1", "title": "something", "list_id": "fakeid1"}`).
		Status(http.StatusOK).
		End()
	NewTest(app).
		Post("/todos/fakeid1/done").
		Header("Content-Type", "application/json").
		Body(`true`).
		Expect(t).
		Status(http.StatusOK).
		Body(`{}`).
		End()
	NewTest(app).
		Get("/todos/fakeid1").
		Expect(t).
		Body(`{"done": true, "id": "fakeid1", "title": "something", "list_id": "fakeid1"}`).
		Status(http.StatusOK).
		End()
}

func TestOneTodoListWithThreeItems(t *testing.T) {
	app := newApp()
	var list TodoList
	items := make([]TodoItem, 3)

	// Create a todo list
	NewTest(app).
		Post("/lists").
		Header("Content-Type", "application/json").
		Body(`{"name": "test234"}`).
		Expect(t).
		Status(http.StatusOK).
		End().JSON(&list)

	// Create 3 todo items
	NewTest(app).
		Post("/todos").
		Header("Content-Type", "application/json").
		Body(`{"title": "milk", "list_id": "` + list.Id + `"}`).
		Expect(t).
		Status(http.StatusOK).
		End().JSON(&items[0])
	NewTest(app).
		Post("/todos").
		Header("Content-Type", "application/json").
		Body(`{"title": "bread", "list_id": "` + list.Id + `"}`).
		Expect(t).
		Status(http.StatusOK).
		End().JSON(&items[1])
	NewTest(app).
		Post("/todos").
		Header("Content-Type", "application/json").
		Body(`{"title": "egg", "list_id": "` + list.Id + `"}`).
		Expect(t).
		Status(http.StatusOK).
		End().JSON(&items[2])

	// Test features of the todo list and the todo items
	NewTest(app).
		Get("/list/" + list.Id).
		Expect(t).
		Body(`{"id": "` + list.Id + `", "name": "test234"}`).
		Status(http.StatusOK).
		End()
	NewTest(app).
		Get("/list/" + list.Id + "/items").
		QueryParams(map[string]string{"len": "10"}).
		Expect(t).
		Body(fmt.Sprintf(`["%v", "%v", "%v"]`, items[0].Id, items[1].Id, items[2].Id)).
		Status(http.StatusOK).
		End()
	NewTest(app).
		Get("/list/" + list.Id + "/items").
		QueryParams(map[string]string{"pos": "1", "len": "2"}).
		Expect(t).
		Body(fmt.Sprintf(`["%v", "%v"]`, items[1].Id, items[2].Id)).
		Status(http.StatusOK).
		End()

	NewTest(app).
		Get("/list/" + list.Id + "/items").
		QueryParams(map[string]string{"done": "false"}).
		Expect(t).
		Body(fmt.Sprintf(`["%v", "%v", "%v"]`, items[0].Id, items[1].Id, items[2].Id)).
		Status(http.StatusOK).
		End()
	NewTest(app).
		Get("/list/" + list.Id + "/items").
		QueryParams(map[string]string{"done": "true"}).
		Expect(t).
		Body(`[]`).
		Status(http.StatusOK).
		End()
	NewTest(app).
		Post("/todos/"+items[1].Id+"/done").
		Header("Content-Type", "application/json").
		Body(`true`).
		Expect(t).
		Status(http.StatusOK).
		Body(`{}`).
		End()
	NewTest(app).
		Get("/list/" + list.Id + "/items").
		QueryParams(map[string]string{"done": "false"}).
		Expect(t).
		Body(fmt.Sprintf(`["%v", "%v"]`, items[0].Id, items[2].Id)).
		Status(http.StatusOK).
		End()
	NewTest(app).
		Get("/list/" + list.Id + "/items").
		QueryParams(map[string]string{"done": "true"}).
		Expect(t).
		Body(fmt.Sprintf(`["%v"]`, items[1].Id)).
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
