package awsar

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

const (
	assumeRoleTimeout = time.Second * 3
)

// WithAssumedRole assumes an AWS IAM role and performs a function as the assumed role.
func WithAssumedRole(
	ctx context.Context,
	sess *session.Session,
	ari *sts.AssumeRoleInput,
	fn func(context.Context, *session.Session) error,
) error {
	// new AWS STS service client
	stsService := sts.New(sess)

	// new context object with a timeout
	assumeRoleContext, assumeRoleContextCancel := context.WithTimeout(ctx, assumeRoleTimeout)
	defer assumeRoleContextCancel()

	// perform STS AssumeRole API call to get temporary credentials for AWS IAM role
	assumeRoleOutput, err := stsService.AssumeRoleWithContext(assumeRoleContext, ari)
	if err != nil {
		return fmt.Errorf("failed to assume role \"%s\": %s", aws.StringValue(ari.RoleArn), err)
	}

	// initialize new AWS session using temporary credentials from assumed role
	assumedRoleSession, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(
			aws.StringValue(assumeRoleOutput.Credentials.AccessKeyId),
			aws.StringValue(assumeRoleOutput.Credentials.SecretAccessKey),
			aws.StringValue(assumeRoleOutput.Credentials.SessionToken),
		),
	})
	if err != nil {
		return fmt.Errorf("failed to initialize new aws session with temporary credentials for assumed role \"%s\": %s", aws.StringValue(ari.RoleArn), err)
	}

	// run the passed function with the assumed role session
	return fn(ctx, assumedRoleSession)
}
