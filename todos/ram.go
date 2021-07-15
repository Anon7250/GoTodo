package todos

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
)

type RAMTodoDB struct {
	jsons            sync.Map
	strLists         sync.Map
	transactionLocks sync.Map
}

func NewRAMTodoList() (*TodoListAPI, error) {
	ramlist := RAMTodoDB{sync.Map{}, sync.Map{}, sync.Map{}}
	return &TodoListAPI{db: &ramlist}, nil
}

func (todo *RAMTodoDB) HasKey(key string) (bool, error) {
	todo.lock(key)
	defer todo.unlock(key)

	_, ok1 := todo.jsons.Load(key)
	_, ok2 := todo.strLists.Load(key)
	return ok1 || ok2, nil
}

func (todo *RAMTodoDB) GetJson(key string, valueOut interface{}) error {
	todo.lock(key)
	defer todo.unlock(key)

	ans, ok := todo.jsons.Load(key)
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "No such DB key: "+key)
	}

	ansBytes, ok := ans.([]byte)
	if !ok {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Sprintf("Invalid json for key '%s': %T: %#v", key, ans, ans),
		)
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
	todo.jsons.Store(key, ans)
	return nil
}

func (todo *RAMTodoDB) GetStringList(key string, valueOut *[]string) error {
	todo.lock(key)
	defer todo.unlock(key)

	rawList, ok := todo.strLists.Load(key)
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "No such DB key: "+key)
	}

	ans, ok := rawList.([]string)
	if !ok {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Sprintf("Invalid list for key '%s': %T: %#v", key, rawList, rawList),
		)
	}

	*valueOut = ans
	return nil
}

func (todo *RAMTodoDB) DoWriteTransaction(t WriteTransaction) error {
	locked_map := make(map[string]bool)
	for key := range t.creates {
		_, locked := locked_map[key]
		if locked {
			continue
		}
		todo.lock(key)
		defer todo.unlock(key)
		locked_map[key] = true
	}
	for key := range t.overwrites {
		_, locked := locked_map[key]
		if locked {
			continue
		}
		todo.lock(key)
		defer todo.unlock(key)
		locked_map[key] = true
	}
	for key := range t.strListAppends {
		_, locked := locked_map[key]
		if locked {
			continue
		}
		todo.lock(key)
		defer todo.unlock(key)
		locked_map[key] = true
	}
	for _, key := range t.strListCreates {
		_, locked := locked_map[key]
		if locked {
			continue
		}
		todo.lock(key)
		defer todo.unlock(key)
		locked_map[key] = true
	}

	for key := range t.creates {
		_, currBytesExists := todo.jsons.Load(key)
		if currBytesExists {
			return fiber.NewError(
				fiber.StatusInternalServerError,
				fmt.Sprintf("Transanction failed: expecting key %v not to exist, but it does", key),
			)
		}
	}
	for _, key := range t.strListCreates {
		_, currBytesExists := todo.strLists.Load(key)
		if currBytesExists {
			return fiber.NewError(
				fiber.StatusInternalServerError,
				fmt.Sprintf("Transanction failed: expecting list %v not to exist, but it does", key),
			)
		}
	}

	oldLists := make(map[string][]string)
	for key := range t.strListAppends {
		list, currBytesExists := todo.strLists.Load(key)
		if !currBytesExists {
			return fiber.NewError(
				fiber.StatusInternalServerError,
				fmt.Sprintf("Transanction failed: expecting list %v to exist, but it does not", key),
			)
		}
		oldLists[key] = list.([]string)
	}

	for key, obj := range t.creates {
		rawJson, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		todo.jsons.Store(key, rawJson)
	}
	for key, obj := range t.overwrites {
		rawJson, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		todo.jsons.Store(key, rawJson)
	}
	for _, key := range t.strListCreates {
		var list = make([]string, 0)
		todo.strLists.Store(key, list)
	}
	for key, appends := range t.strListAppends {
		list := oldLists[key]
		list = append(list, appends...)
		todo.strLists.Store(key, list)
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
