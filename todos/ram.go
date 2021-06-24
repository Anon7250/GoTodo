package todos

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
)

type RAMTodoDB map[string][]byte

func NewRAMTodoList() (*TodoListAPI, error) {
	ramlist := make(RAMTodoDB)
	return &TodoListAPI{db: &ramlist}, nil
}

func (todo *RAMTodoDB) HasKey(key string) (bool, error) {
	_, ok := (*todo)[key]
	return ok, nil
}

func (todo *RAMTodoDB) GetJson(key string, valueOut interface{}) error {
	ans, ok := (*todo)[key]
	if !ok {
		return errors.New("Invalid DB key: " + key)
	}
	return json.Unmarshal(ans, valueOut)
}

func (todo *RAMTodoDB) SetJson(key string, value interface{}) error {
	ans, err := json.Marshal(value)
	if err != nil {
		return err
	}
	(*todo)[key] = ans
	return nil
}

func (todo *RAMTodoDB) ListJsons(keyPrefix string, valuesOut interface{}) error {
	var ans [][]byte
	for key, val := range *todo {
		if strings.HasPrefix(key, keyPrefix) {
			ans = append(ans, val)
		}
	}
	rawJsonBuilder := [][]byte{
		[]byte("["),
		bytes.Join(ans, []byte(",")),
		[]byte("]"),
	}
	rawJson := bytes.Join(rawJsonBuilder, []byte(""))
	return json.Unmarshal(rawJson, valuesOut)
}
