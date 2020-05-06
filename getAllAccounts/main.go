package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Response struct {
	Response []account `json:"response"`
}
type account struct {
	UserName   string `json:"user_name" dynamodbav:"user_name" validate:"required"`
	FullName   string `json:"full_name" dynamodbav:"full_name" validate:"required"`
	ProfileUrl string `json:"profile_url" dynamodbav:"profile_url"`
}

func Session_init() (*dynamodb.DynamoDB, error) {
	region := os.Getenv("AWS_REGION")
	// Initialize a session
	if session, err := session.NewSession(&aws.Config{
		Region: &region,
	}); err != nil {
		fmt.Println(fmt.Sprintf("Failed to initialize a session to AWS: %s", err.Error()))
		return nil, err
		// Create DynamoDB client
	} else {
		return dynamodb.New(session), nil
	}
}
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var (
		err       error
		svc       *dynamodb.DynamoDB
		tableName = aws.String(os.Getenv("IAM_TABLE_NAME"))
		accounts  []account
	)
	// Intitalize a client session to dynamodb
	svc, err = Session_init()
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Session init erro " + err.Error(), StatusCode: 500}, nil
	}
	params := &dynamodb.ScanInput{
		TableName: tableName,
	}
	result, err := svc.Scan(params)
	if err != nil {
		// Status Bad Request
		return events.APIGatewayProxyResponse{
			Body:       "DynamoDB Query API call failed: " + err.Error(),
			StatusCode: 500,
		}, nil
	}

	for _, i := range result.Items {
		user_account := account{}
		if err := dynamodbattribute.UnmarshalMap(i, &user_account); err != nil {
			return events.APIGatewayProxyResponse{
				Body:       "Got error unmarshalling:" + err.Error(),
				StatusCode: 500,
			}, nil
		}
		accounts = append(accounts, user_account)
	}

	body, _ := json.Marshal(&Response{
		Response: accounts,
	})
	return events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: 200,
	}, nil

}

func main() {
	lambda.Start(Handler)
}
