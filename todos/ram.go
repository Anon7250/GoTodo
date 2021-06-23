package todos

import (
	"errors"
	"strings"
)

type RAMTodoDB map[string][]byte

func NewRAMTodoList() (*TodoList, error) {
	ramlist := make(RAMTodoDB, 0)
	return &TodoList{db: &ramlist}, nil
}

func (todo *RAMTodoDB) GetJson(key string) ([]byte, error) {
	ans, ok := (*todo)[key]
	if !ok {
		return nil, errors.New("Invalid DB key: " + key)
	}
	return ans, nil
}

func (todo *RAMTodoDB) SetJson(key string, json []byte) error {
	(*todo)[key] = json
	return nil
}

func (todo *RAMTodoDB) ListJsons(keyPrefix string) ([][]byte, error) {
	var ans [][]byte
	for key, val := range *todo {
		if strings.HasPrefix(key, keyPrefix) {
			ans = append(ans, val)
		}
	}
	return ans, nil
}
