package worker

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// HandlerFunc is used to define the Handler that is run on for each message
type HandlerFunc func(msg *sqs.Message) error

func (f HandlerFunc) HandleMessage(msg *sqs.Message) error {
	return f(msg)
}

// Handler interface
type Handler interface {
	HandleMessage(msg *sqs.Message) error
}

type InvalidEventError struct {
	event string
	msg   string
}

func (e InvalidEventError) Error() string {
	return fmt.Sprintf("[Invalid Event: %s] %s", e.event, e.msg)
}

func NewInvalidEventError(event, msg string) InvalidEventError {
	return InvalidEventError{event: event, msg: msg}
}

type Config struct {
	// what is the queue url we are connecting to, Defaults to empty
	QueueURL string
	// The maximum number of messages to return. Amazon SQS never returns more messages
	// than this value (however, fewer messages might be returned). Valid values
	// are 1 to 10.
	MaxNumberOfMessage int64
	// The duration (in seconds) for which the call waits for a message to arrive
	// in the queue before returning. If a message is available, the call returns
	// sooner than WaitTimeSeconds.
	WaitTimeSecond int64

	Log            LoggerIFace
	LoggingEnabled bool
}

// Start starts the polling and will continue polling till the application is forcibly stopped
func (c *Config) Start(svc *sqs.SQS, h Handler) {
	for {
		c.Log.Debug("worker: Start Polling")

		params := &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(c.QueueURL), // Required
			MaxNumberOfMessages: aws.Int64(c.MaxNumberOfMessage),
			MessageAttributeNames: []*string{
				aws.String("All"), // Required
			},
			WaitTimeSeconds: aws.Int64(c.WaitTimeSecond),
		}

		resp, err := svc.ReceiveMessage(params)

		if err != nil {
			c.Log.Error(err)
			continue
		}

		if len(resp.Messages) > 0 {
			c.run(svc, h, resp.Messages)
		}
	}
}

// poll launches goroutine per received message and wait for all message to be processed
func (c *Config) run(svc *sqs.SQS, h Handler, messages []*sqs.Message) {
	numMessages := len(messages)
	c.Log.Info(fmt.Sprintf("worker: Received %d messages", numMessages))

	var wg sync.WaitGroup
	wg.Add(numMessages)

	for i := range messages {
		go func(m *sqs.Message) {
			// launch goroutine
			defer wg.Done()

			if err := c.handleMessage(svc, m, h); err != nil {
				if c.LoggingEnabled {
					c.Log.Error(err.Error())
				}
			}
		}(messages[i])
	}

	wg.Wait()
}

func (c *Config) handleMessage(svc *sqs.SQS, m *sqs.Message, h Handler) error {
	var err error
	err = h.HandleMessage(m)

	if _, ok := err.(InvalidEventError); ok {
		c.Log.Error(err.Error())
	} else if err != nil {
		return err
	}

	params := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.QueueURL), // Required
		ReceiptHandle: m.ReceiptHandle,        // Required
	}
	_, err = svc.DeleteMessage(params)

	if err != nil {
		return err
	}
	c.Log.Debug(fmt.Sprintf("worker: deleted message from queue: %s", aws.StringValue(m.ReceiptHandle)))

	return nil
}
