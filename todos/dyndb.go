// Use AWS Dynamodb as storage
package todos

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	dyndb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dyndbTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gofiber/fiber/v2"
)

const DefaultTableName = "GoTodo1"
const TableKey = "key"
const TableJsonField = "rawJson"
const TableStrListField = "strList"
const StrListCreatedMarker = "<CREATED>"
const ConditionKeyDoesntExist = "attribute_not_exists(#key)"
const ConditionKeyExists = "attribute_exists(#key)"
const AppendToStrListExpr = "SET strList = list_append(strList, :AppendItems)"
const RenameNewItems = ":AppendItems"
const RenameTableKey = "#key"
const RequestTokenSize = 36

type DynDBTodoDB struct {
	AwsConfig *aws.Config
	DB        *dyndb.Client
	Table     string
}

func NewDynDBTodoList(table string) (*TodoListAPI, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	db := dyndb.NewFromConfig(cfg)
	if len(table) == 0 {
		table = DefaultTableName
	}

	impl := DynDBTodoDB{&cfg, db, table}
	return &TodoListAPI{db: &impl}, nil
}

func (todo *DynDBTodoDB) getItem(key string) (*dyndb.GetItemOutput, error) {
	input := dyndb.GetItemInput{
		Key: map[string]dyndbTypes.AttributeValue{
			TableKey: &dyndbTypes.AttributeValueMemberS{Value: key},
		},
		TableName: aws.String(todo.Table),
	}
	return todo.DB.GetItem(context.TODO(), &input)
}

func (todo *DynDBTodoDB) HasKey(key string) (bool, error) {
	errMsg := "Cannot read from database: " + key + ". "
	result, err := todo.getItem(key)
	if err != nil {
		return false, fiber.NewError(fiber.StatusInternalServerError, errMsg+err.Error())
	}
	hasKey := result.Item != nil
	return hasKey, nil
}

func (todo *DynDBTodoDB) GetJson(key string, valueOut interface{}) error {
	result, err := todo.getItem(key)
	if err != nil {
		return err
	}
	if result.Item == nil {
		return fiber.NewError(fiber.StatusNotFound)
	}

	errMsg := "Cannot read JSON from database: " + key + ". "
	rawJson, ok := result.Item[TableJsonField]
	if !ok {
		return fiber.NewError(fiber.StatusInternalServerError, errMsg)
	}

	err = attributevalue.Unmarshal(rawJson, valueOut)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errMsg+err.Error())
	}
	return nil
}

func (todo *DynDBTodoDB) GetStringList(key string, valueOut *[]string) error {
	result, err := todo.getItem(key)
	if err != nil {
		return err
	}
	if result.Item == nil {
		return fiber.NewError(fiber.StatusNotFound)
	}

	errMsg := "Cannot read a list from database: " + key + ". "
	rawStrList, ok := result.Item[TableStrListField]
	if !ok {
		return fiber.NewError(fiber.StatusInternalServerError, errMsg)
	}

	err = attributevalue.Unmarshal(rawStrList, valueOut)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errMsg)
	}
	return nil
}

func (todo *DynDBTodoDB) DoWriteTransaction(t WriteTransaction) error {
	errMsg := "Failed to write to database: "
	uuidStr, err := GetUUIDImpl()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errMsg+"Cannot generate UUID for transaction requests")
	}

	transactions := make([]dyndbTypes.TransactWriteItem, 0)
	createsKeys := make(map[string]bool)
	createsStrLists := make(map[string]bool)
	for key := range t.creates {
		createsKeys[key] = true
	}
	for _, key := range t.strListCreates {
		createsKeys[key] = true
		createsStrLists[key] = true
	}
	for key := range createsKeys {
		values := map[string]dyndbTypes.AttributeValue{
			TableKey: &dyndbTypes.AttributeValueMemberS{Value: key},
		}
		setJson, hasSetJson := t.creates[key]
		_, createStrList := createsStrLists[key]
		if hasSetJson {
			rawJson, err := attributevalue.Marshal(setJson)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, errMsg+err.Error())
			}
			values[TableJsonField] = rawJson
		}
		if createStrList {
			rawJson, err := attributevalue.Marshal([]string{})
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, errMsg+err.Error())
			}
			values[TableStrListField] = rawJson
		}
		item := dyndbTypes.TransactWriteItem{
			Put: &dyndbTypes.Put{
				ExpressionAttributeNames: map[string]string{RenameTableKey: TableKey},
				Item:                     values,
				TableName:                aws.String(todo.Table),
				ConditionExpression:      aws.String(ConditionKeyDoesntExist),
			},
		}
		transactions = append(transactions, item)
	}
	for key, setJson := range t.overwrites {
		rawJson, err := attributevalue.Marshal(setJson)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, errMsg+err.Error())
		}
		item := dyndbTypes.TransactWriteItem{
			Put: &dyndbTypes.Put{
				ExpressionAttributeNames: map[string]string{RenameTableKey: TableKey},
				Item: map[string]dyndbTypes.AttributeValue{
					TableKey:       &dyndbTypes.AttributeValueMemberS{Value: key},
					TableJsonField: rawJson,
				},
				TableName: aws.String(todo.Table),
			},
		}
		transactions = append(transactions, item)
	}
	for key, strs := range t.strListAppends {
		rawJson, err := attributevalue.Marshal(strs)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, errMsg+err.Error())
		}
		item := dyndbTypes.TransactWriteItem{
			Update: &dyndbTypes.Update{
				ExpressionAttributeNames: map[string]string{RenameTableKey: TableKey},
				ExpressionAttributeValues: map[string]dyndbTypes.AttributeValue{
					RenameNewItems: rawJson,
				},
				Key: map[string]dyndbTypes.AttributeValue{
					TableKey: &dyndbTypes.AttributeValueMemberS{Value: key},
				},
				TableName:           aws.String(todo.Table),
				ConditionExpression: aws.String(ConditionKeyExists),
				UpdateExpression:    aws.String(AppendToStrListExpr),
			},
		}
		transactions = append(transactions, item)
	}

	if len(uuidStr) > RequestTokenSize {
		uuidStr = uuidStr[len(uuidStr)-RequestTokenSize:]
	}
	input := dyndb.TransactWriteItemsInput{
		TransactItems:      transactions,
		ClientRequestToken: aws.String(uuidStr),
	}
	_, err = todo.DB.TransactWriteItems(context.TODO(), &input)
	if err != nil {
		fmt.Printf("Error: %v", err.Error())
		return fiber.NewError(fiber.StatusInternalServerError, errMsg+err.Error())
	}
	return nil
}
