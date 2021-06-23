// Use AWS Dynamodb as storage
package todos

import (
	"context"
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

func NewDynDBTodoList() (*TodoList, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	db := dynamodb.NewFromConfig(cfg)

	impl := DynDBTodoDB{&cfg, db}
	return &TodoList{db: &impl}, nil
}

// TODO: This is very expensive for AWS. Get rid of it
func (todo *DynDBTodoDB) ListJsons(keyPrefix string) ([][]byte, error) {
	var table = "GoTodo1"
	var projectExpr = "id"
	scanInput := dynamodb.ScanInput{
		TableName:            &table,
		ProjectionExpression: &projectExpr,
	}
	output, err := todo.DB.Scan(context.TODO(), &scanInput)
	if err != nil {
		return nil, err
	}
	log.Printf("AWS Scanned %d items and returned %d", output.ScannedCount, output.Count)

	return nil, nil
}

func (todo *DynDBTodoDB) GetJson(key string) ([]byte, error) {
	return nil, fiber.NewError(fiber.StatusNotFound)
}

func (todo *DynDBTodoDB) SetJson(key string, json []byte) error {
	return nil
}
