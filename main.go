package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// CONFIGURATION
const (
	InstanceID = "i-005a50b03a5dfffe9"
	TargetType = types.InstanceTypeT3Small
	Region     = "us-east-1"
)

func HandleRequest(ctx context.Context) (string, error) {
	
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(Region))
	if err != nil {
		return "", fmt.Errorf("unable to load SDK config, %v", err)
	}

	client := ec2.NewFromConfig(cfg)


	fmt.Printf("Stopping instance %s...\n", InstanceID)
	_, err = client.StopInstances(ctx, &ec2.StopInstancesInput{
		InstanceIds: []string{InstanceID},
	})
	if err != nil {
		return "", fmt.Errorf("failed to stop instance: %v", err)
	}

	t 
	fmt.Println("Waiting for instance to stop...")
	waiter := ec2.NewInstanceStoppedWaiter(client)
	err = waiter.Wait(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{InstanceID},
	}, 5*time.Minute)
	if err != nil {
		return "", fmt.Errorf("error waiting for instance to stop: %v", err)
	}

	// 4. Modify the Instance Type (The Resize)
	fmt.Printf("Resizing instance to %s...\n", TargetType)
	_, err = client.ModifyInstanceAttribute(ctx, &ec2.ModifyInstanceAttributeInput{
		InstanceId: aws.String(InstanceID),
		InstanceType: &types.AttributeValue{
			Value: aws.String(string(TargetType)),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to resize instance: %v", err)
	}

	fmt.Println("Starting instance...")
	_, err = client.StartInstances(ctx, &ec2.StartInstancesInput{
		InstanceIds: []string{InstanceID},
	})
	if err != nil {
		return "", fmt.Errorf("failed to start instance: %v", err)
	}

	return fmt.Sprintf("Successfully resized %s to %s and started it.", InstanceID, TargetType), nil
}

func main() {
	lambda.Start(HandleRequest)
}
