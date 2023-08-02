package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	apiv1 "github.com/acorn-io/runtime/pkg/apis/api.acorn.io/v1"
	internalv1 "github.com/acorn-io/runtime/pkg/apis/internal.acorn.io/v1"
	kclient "github.com/acorn-io/runtime/pkg/k8sclient"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sClient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	stackNameEnvKey = "ACORN_EXTERNAL_ID"
	acornEventKey   = "ACORN_EVENT"
	acornNameKey    = "ACORN_NAME"
	acornProjectKey = "ACORN_PROJECT"
	MESSAGE         = "CFN Stack: %s provisioning, %d/%d/%d(ready/failed/total)\n"
)

type ServerConfig struct {
	StackName     string
	Context       context.Context
	CfnClient     *cloudformation.Client
	AwsConfig     aws.Config
	KClient       k8sClient.WithWatch
	Transitioning bool
}

func main() {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logrus.Fatal(err)
	}

	k8sClient, err := kclient.Default()
	if err != nil {
		logrus.Fatal(err)
	}

	awsClient := cloudformation.NewFromConfig(cfg)

	sc := &ServerConfig{
		StackName: os.Getenv(stackNameEnvKey),
		Context:   ctx,
		CfnClient: awsClient,
		AwsConfig: cfg,
		KClient:   k8sClient,
	}

	if sc.StackName == "" {
		logrus.Fatalf("Missing %s environment variable pointing to Cloud Formation StackName", stackNameEnvKey)
	}

	if err := sc.watchStack(); err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("Stack %s is ready", sc.StackName)

}

func (sc *ServerConfig) watchStack() error {

	for {
		select {
		case <-sc.Context.Done():
			return nil
		default:
			stackResources, err := sc.CfnClient.DescribeStackResources(sc.Context, &cloudformation.DescribeStackResourcesInput{
				StackName: aws.String(sc.StackName),
			})
			if err != nil {
				logrus.Error(err)
				time.Sleep(10 * time.Second)
				continue
			}

			totalCount := len(stackResources.StackResources)
			var readyCount int
			var failedResources int
			for _, resource := range stackResources.StackResources {
				if resource.ResourceStatus == types.ResourceStatusCreateComplete || resource.ResourceStatus == types.ResourceStatusUpdateComplete {
					readyCount++
				}
				if resource.ResourceStatus == types.ResourceStatusCreateFailed || resource.ResourceStatus == types.ResourceStatusUpdateFailed {
					failedResources++
				}
			}
			if readyCount < totalCount || failedResources > 0 {
				sc.Transitioning = true
			}

			if readyCount == totalCount && sc.Transitioning {
				sc.Transitioning = false
				sc.emit(readyCount, failedResources, totalCount)
				continue
			} else if readyCount == totalCount {
				continue
			}

			sc.emit(readyCount, failedResources, totalCount)
			time.Sleep(30 * time.Second)
		}
	}
}

func (sc *ServerConfig) emit(r, f, total int) {
	if err := sc.emitEvent(r, f, total); err != nil {
		logrus.Error(err)
	}

	if err := sc.logStdOut(r, f, total); err != nil {
		logrus.Error(err)
	}
}

func (sc *ServerConfig) emitEvent(r, f, total int) error {
	e := &apiv1.Event{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Event",
			APIVersion: "api.acorn.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    os.Getenv(acornProjectKey),
			GenerateName: "se-",
		},

		Type:        getEventType(sc.Transitioning),
		AppName:     os.Getenv(acornNameKey),
		Severity:    "info",
		Description: fmt.Sprintf(MESSAGE, sc.StackName, r, f, total),
		Resource: &internalv1.EventResource{
			Kind: "app",
			Name: os.Getenv(acornNameKey),
		},
	}
	return sc.KClient.Create(sc.Context, e)
}

func (sc *ServerConfig) logStdOut(r, f, total int) error {
	fmt.Printf(MESSAGE, sc.StackName, r, f, total)
	return nil
}

func getEventType(t bool) string {
	event := os.Getenv(acornEventKey)
	if !t {
		return fmt.Sprintf("Service%sd", strings.ToTitle(event))
	}

	eventPhrase := map[string]string{
		"update": "ServiceUpdating",
		"create": "ServiceCreating",
		"delete": "ServiceDeleting",
	}
	return eventPhrase[event]
}
