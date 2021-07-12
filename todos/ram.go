package todos

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
)

type RAMTodoDB struct {
	data             sync.Map
	transactionLocks sync.Map
}

func NewRAMTodoList() (*TodoListAPI, error) {
	ramlist := RAMTodoDB{sync.Map{}, sync.Map{}}
	return &TodoListAPI{db: &ramlist}, nil
}

func (todo *RAMTodoDB) HasKey(key string) (bool, error) {
	todo.lock(key)
	defer todo.unlock(key)

	_, ok := todo.data.Load(key)
	return ok, nil
}

func (todo *RAMTodoDB) GetJson(key string, valueOut interface{}) error {
	todo.lock(key)
	defer todo.unlock(key)

	ans, ok := todo.data.Load(key)
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "No such DB key: "+key)
	}

	ansBytes, ok := ans.([]byte)
	if !ok {
		return fmt.Errorf("invalid json for key '%s': %T: %#v", key, ans, ans)
	}
	return json.Unmarshal(ansBytes, valueOut)
}

func (todo *RAMTodoDB) SetJson(key string, value interface{}) error {
	todo.lock(key)
	defer todo.unlock(key)

	ans, err := json.Marshal(value)
	if err != nil {
		return err
	}
	todo.data.Store(key, ans)
	return nil
}

func (todo *RAMTodoDB) TransactSetJsons(writes map[string]interface{}, conditions map[string]interface{}) error {
	for key, obj := range writes {
		todo.lock(key)
		defer todo.unlock(key)

		condition, exists := conditions[key]
		if exists && condition != obj {
			return fmt.Errorf("transanction failed for key: %v because %#v != %#v)", key, obj, condition)
		}

		rawJson, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		todo.data.Store(key, rawJson)
	}
	return nil
}

func (todo *RAMTodoDB) lock(key string) {
	mutex, _ := todo.transactionLocks.LoadOrStore(key, &sync.Mutex{})
	mutex.(*sync.Mutex).Lock()
}

func (todo *RAMTodoDB) unlock(key string) {
	mutex, ok := todo.transactionLocks.Load(key)
	if !ok {
		return
	}

	mutex.(*sync.Mutex).Unlock()
}
