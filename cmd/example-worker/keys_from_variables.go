package main

import (
	"fmt"

	"github.com/Sef1995/go-sqs-poller/worker"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/sqs"
)

var (
	awsAccessKey string
	awsSecretKey string
	awsRegion    string
)

func main() {

	// create a config
	awsConfig := &aws.Config{
		Credentials: credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, ""),
		Region:      aws.String(awsRegion),
	}

	// create the new client with the aws_config passed in
	svc, url := worker.NewSQSClient("go-webhook-queue-test", awsConfig)

	cfg := worker.Config{
		QueueURL:           url,
		MaxNumberOfMessage: 10,
		WaitTimeSecond:     20,
		LoggingEnabled:     true,
	}

	cfg.Start(svc, worker.HandlerFunc(func(msg *sqs.Message) error {
		fmt.Println(aws.StringValue(msg.Body))
		return nil
	}))
}
