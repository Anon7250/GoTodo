module github.com/Anon7250/gotodo

go 1.16

require (
	github.com/aws/aws-sdk-go-v2 v1.7.1
	github.com/aws/aws-sdk-go-v2/config v1.4.1
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.1.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.4.1
	github.com/gofiber/fiber/v2 v2.10.0
	github.com/google/uuid v1.2.0
	github.com/prashantv/gostub v1.0.0 // indirect
	github.com/steinfletcher/apitest v1.5.10 // indirect
)

replace github.com/Anon7250/gotodo/todos => ./todos
