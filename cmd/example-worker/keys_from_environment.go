package main

import (
	"fmt"

	"github.com/Sef1995/go-sqs-poller/worker"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func main() {
	// Make sure that AWS_SDK_LOAD_CONFIG=true is defined as an environment variable before running the application
	// like this

	// create the new client and return the url
	svc1, url1 := worker.NewSQSClient("go-webhook-queue-test")

	cfg := worker.Config{
		QueueURL:           url1,
		MaxNumberOfMessage: 10,
		WaitTimeSecond:     20,
		LoggingEnabled:     true,
	}

	cfg.Start(svc1, worker.HandlerFunc(func(msg *sqs.Message) error {
		fmt.Println(aws.StringValue(msg.Body))
		return nil
	}))
}
