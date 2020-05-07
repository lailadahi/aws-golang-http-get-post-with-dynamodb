# Serverless-golang http Get and Post with DynamoDB resource example 

## Medium post
[can be found here](https://medium.com/@lailadahi/serverless-service-on-aws-lambda-go-sdk-dynamodb-7551f943d1dc)

## prerequisites
    - serverless 
    - go
    - aws cli  

## Use this repo as a servrless template 
    sls create --template-url https://github.com/lailadahi/aws-golang-http-get-post-with-dynamodb -n accounts-service

## Deploy 
    $ cd accounts-service
    $ serverless deploy 

#### Create an account  
    $ curl -H "Content-Type: application/json" -X POST -d @data.json  https://<your_lamda_function_hostname>/dev/accounts/
#### List all accounts 
    $ curl -X GET https://<your_lamda_function_hostname>/dev/accounts/
#### Get a specific account by the username  
    $ curl -X GET https://<your_lamda_function_hostname>/dev/accounts/{user_name}
