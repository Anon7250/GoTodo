package main

import (
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
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
    Body(`[{"title": "something"}]`).
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
