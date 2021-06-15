// Use AWS Dynamodb as storage
package todos

import (
	"context"
	"log"
	"github.com/gofiber/fiber/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynDBTodoList struct {
	AwsConfig *aws.Config
	DB *dynamodb.Client
}

func NewDynDBTodoList() (*TodoList, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	db := dynamodb.NewFromConfig(cfg)

	impl := DynDBTodoList{&cfg, db}
  return &TodoList{impl: &impl}, nil
}

func (todo *DynDBTodoList) GetAll() ([]string, error) {
	var table = "GoTodo1"
	var projectExpr = "id"
	scanInput := dynamodb.ScanInput {
		TableName: &table,
		ProjectionExpression: &projectExpr,
	}
	output, err := todo.DB.Scan(context.TODO(), &scanInput)
	if err != nil {
		return nil, err
	}
	log.Printf("AWS Scanned %d items and returned %d", output.ScannedCount, output.Count)
	return []string {}, nil
}

func (todo *DynDBTodoList) GetById(id string) (*TodoItem, error) {
  return nil, fiber.NewError(fiber.StatusNotFound)
}

func (todo *DynDBTodoList) AddTodo(item TodoItem) error {
	return nil
}
