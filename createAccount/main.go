package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"gopkg.in/go-playground/validator.v9"
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
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}, nil
	}
	// BodyRequest will be used to take the json response from client and build it
	bodyRequest := account{}

	// Unmarshal the json, return 404 if error
	err = json.Unmarshal([]byte(request.Body), &bodyRequest)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 404}, nil
	}
	// Validate that the request has all the required fields
	var validate *validator.Validate
	validate = validator.New()
	err = validate.Struct(&bodyRequest)
	if err != nil {
		// Status Bad Request
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: 400,
		}, nil
	}
	// Create an map of DynamoDB attributes for the item that's about to be added
	item, err := dynamodbattribute.MarshalMap(&bodyRequest)
	if err != nil {
		fmt.Println("Got error calling MarshalMap:")
		fmt.Println(err.Error())
		// Status Bad Request
		return events.APIGatewayProxyResponse{
			Body:       "MarshalMap error: " + err.Error(),
			StatusCode: 400,
		}, nil
	}
	// Construct the putItem input
	params := &dynamodb.PutItemInput{
		Item:                item,
		TableName:           tableName,
		ConditionExpression: aws.String("attribute_not_exists(Username)"), //If item exists, thow an error
	}
	// PutItem to the DynamoDB specified above
	if _, err := svc.PutItem(params); err != nil {
		// Item exists
		if err.(awserr.Error).Code() == dynamodb.ErrCodeConditionalCheckFailedException {
			return events.APIGatewayProxyResponse{
				Body:       fmt.Sprintf("A profile with username=%s already exists. Error: %s", bodyRequest.UserName, err.Error()),
				StatusCode: 500,
			}, nil
		}
		// Status Internal Server Error
		return events.APIGatewayProxyResponse{
			Body:       "svc.PutItem error: " + err.Error(),
			StatusCode: 500,
		}, nil
	} else {
		body, _ := json.Marshal(&bodyRequest)
		// Status OK
		return events.APIGatewayProxyResponse{
			Body:       string(body),
			StatusCode: 201,
		}, nil
	}
}

func main() {
	lambda.Start(Handler)
}
