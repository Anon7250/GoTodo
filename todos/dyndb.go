// Use AWS Dynamodb as storage
package todos

import (
	"context"
	"encoding/json"
	"log"

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

// TODO: This is very expensive for AWS. Get rid of it
func (todo *DynDBTodoDB) ListJsons(keyPrefix string, valuesOut interface{}) error {
	var table = "GoTodo1"
	var projectExpr = "id"
	scanInput := dynamodb.ScanInput{
		TableName:            &table,
		ProjectionExpression: &projectExpr,
	}
	output, err := todo.DB.Scan(context.TODO(), &scanInput)
	if err != nil {
		return err
	}
	log.Printf("AWS Scanned %d items and returned %d", output.ScannedCount, output.Count)

	return json.Unmarshal([]byte("[]"), valuesOut)
}

func (todo *DynDBTodoDB) GetJson(key string, valueOut interface{}) error {
	return fiber.NewError(fiber.StatusNotFound)
}

func (todo *DynDBTodoDB) SetJson(key string, value interface{}) error {
	return nil
}
