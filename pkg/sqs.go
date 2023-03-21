package bomber

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type SqsWorker struct {
	queueName string
	queueUrl  *string
	Client    *sqs.Client
}

func NewSqsWorker(ctx context.Context, queueName string) *SqsWorker {
	s := SqsWorker{
		queueName: queueName,
	}
	client, err := s.getSqsClient(ctx)
	if err != nil {
		panic(err)
	} else {
		s.Client = client
		// Get URL of queue
		gQInput := &sqs.GetQueueUrlInput{
			QueueName: &queueName,
		}

		queueUrl, err := s.Client.GetQueueUrl(ctx, gQInput)
		if err != nil {
			panic(err)
		}
		s.queueUrl = queueUrl.QueueUrl

	}
	return &s
}

func (s *SqsWorker) getSqsClient(ctx context.Context) (*sqs.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get sqs client, configuration error %v", err)
	}

	return sqs.NewFromConfig(cfg), nil

}

func (s *SqsWorker) SendMessage(ctx context.Context, msgBody string, attributes map[string]types.MessageAttributeValue) (*string, error) {
	sMInput := &sqs.SendMessageInput{
		MessageBody: &msgBody,
		QueueUrl:    s.queueUrl,
		//DelaySeconds:            10,
		MessageAttributes: attributes,
		//MessageDeduplicationId:  new(string),
		//MessageGroupId:          new(string),
		//MessageSystemAttributes: map[string]types.MessageSystemAttributeValue{},
	}

	resp, err := s.Client.SendMessage(ctx, sMInput)

	if err != nil {
		return nil, err
	}

	return resp.MessageId, nil
}

const MaxNumberOfMessagesPolled = 10
const VisibilityTimeout = 30

func (s *SqsWorker) Receive(ctx context.Context) ([]types.Message, error) {
	sqsReceiveConfig := &sqs.ReceiveMessageInput{
		QueueUrl:            s.queueUrl,
		MaxNumberOfMessages: MaxNumberOfMessagesPolled,
		VisibilityTimeout:   VisibilityTimeout,
	}

	msgResult, err := s.Client.ReceiveMessage(ctx, sqsReceiveConfig)
	if err != nil {
		return nil, fmt.Errorf("could not receive SQS message, %v", err)
	}

	return msgResult.Messages, nil
}

func (s *SqsWorker) Delete(ctx context.Context, msg types.Message) error {
	_, err := s.Client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      s.queueUrl,
		ReceiptHandle: msg.ReceiptHandle,
	})
	if err != nil {
		return fmt.Errorf("could not delete SQS message, %v", err)
	}

	return nil
}
