// Use AWS Dynamodb as storage
package todos

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gofiber/fiber/v2"
)

type DynDBTodoDB struct {
	AwsConfig *aws.Config
	DB        *dynamodb.Client
}

func NewDynDBTodoList() (*TodoListAPI, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	db := dynamodb.NewFromConfig(cfg)

	impl := DynDBTodoDB{&cfg, db}
	return &TodoListAPI{db: &impl}, nil
}

func (todo *DynDBTodoDB) HasKey(key string) (bool, error) {
	return false, nil
}

func (todo *DynDBTodoDB) GetJson(key string, valueOut interface{}) error {
	return fiber.NewError(fiber.StatusNotFound)
}

func (todo *DynDBTodoDB) SetJson(key string, value interface{}) error {
	return nil
}

func (todo *DynDBTodoDB) TransactSetJsons(writes map[string]interface{}, conditions map[string]interface{}) error {
	return nil
}
