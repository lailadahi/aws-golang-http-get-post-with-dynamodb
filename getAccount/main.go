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
	)
	// Intitalize a client session to dynamodb
	svc, err = Session_init()
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Session init erro " + err.Error(), StatusCode: 500}, nil
	}

	// BodyRequest will be used to take the json response from client and build it
	username := request.PathParameters["username"]
	profile := account{
		UserName: username,
	}
	// build the GetItemInput struct
	params := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"user_name": {
				S: aws.String(username),
			},
		},
		TableName: tableName,
	}
	// GetItem from DynamoDB
	if result, err := svc.GetItem(params); err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "svc.GettItem error: " + err.Error(),
			StatusCode: 500,
		}, nil
	} else {
		profile = account{}
		err = dynamodbattribute.UnmarshalMap(result.Item, &profile)
		if err != nil {
			return events.APIGatewayProxyResponse{
				Body:       "UnmarshalMap error: " + err.Error(),
				StatusCode: 500,
			}, nil
		}
		// If item doesn't exist, DynamoDB returns an empty item, not nil
		if profile.UserName != username {
			return events.APIGatewayProxyResponse{
				Body:       "Username not found: " + username,
				StatusCode: 404,
			}, nil

		}
		// Status OK
		body, _ := json.Marshal(&profile)
		return events.APIGatewayProxyResponse{
			Body:       string(body),
			StatusCode: 200,
		}, nil
	}

}

func main() {
	lambda.Start(Handler)
}
