package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/adrianosela/awsar"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/sts"
)

func main() {
	err := awsar.WithAssumedRole(
		context.Background(),
		session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		})),
		&sts.AssumeRoleInput{
			RoleArn:         aws.String("arn:aws:iam::123456789012:role/CloudWatchWriteRole"),
			ExternalId:      aws.String("super-secret-external-id"),
			RoleSessionName: aws.String(fmt.Sprintf("%d", time.Now().UnixNano())),
			DurationSeconds: aws.Int64(900),
		},
		func(ctx context.Context, sess *session.Session) error {
			// new AWS CloudWatch Logs client
			cw := cloudwatchlogs.New(sess)

			// new context with timeout for AWS CloudWatch PutLogEvents
			putLogEventsContext, putLogEventsContextCancel := context.WithTimeout(ctx, time.Second*3)
			defer putLogEventsContextCancel()

			// perform PutLogEvents operation
			_, err := cw.PutLogEventsWithContext(putLogEventsContext, &cloudwatchlogs.PutLogEventsInput{
				LogGroupName:  aws.String("my-log-group"),
				LogStreamName: aws.String("my-log-stream"),
				LogEvents: []*cloudwatchlogs.InputLogEvent{{
					Timestamp: aws.Int64(time.Now().UnixMilli()),
					Message:   aws.String("my log message!"),
				}},
			})
			if err != nil {
				return fmt.Errorf("failed to put log events: %s", err)
			}

			// success!
			return nil
		},
	)
	if err != nil {
		log.Fatalf("failed to push logs with assumed role: %s", err)
	}
}
