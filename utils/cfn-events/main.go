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
)

type StackWatcher struct {
	Context   context.Context
	CfnClient *cloudformation.Client
	KClient   k8sClient.WithWatch
	Stack     *Stack
}

type Stack struct {
	Name            string
	DeletedCount    int
	ReadyCount      int
	FailedResources int
	TotalCount      int
	Transitioning   bool
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

	sw := &StackWatcher{
		Stack: &Stack{
			Name: os.Getenv(stackNameEnvKey),
		},
		Context:   ctx,
		CfnClient: awsClient,
		KClient:   k8sClient,
	}

	if sw.Stack.Name == "" {
		logrus.Fatalf("Missing %s environment variable pointing to Cloud Formation StackName", stackNameEnvKey)
	}

	if err := sw.watchStack(); err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("Stack %s is ready", sw.Stack.Name)

}

func (sw *StackWatcher) watchStack() error {

	for {
		select {
		case <-sw.Context.Done():
			return nil
		default:
			stackResources, err := sw.CfnClient.DescribeStackResources(sw.Context, &cloudformation.DescribeStackResourcesInput{
				StackName: aws.String(sw.Stack.Name),
			})
			if err != nil {
				logrus.Error(err)
				time.Sleep(10 * time.Second)
				continue
			}

			sw.Stack.ReadyCount = 0
			sw.Stack.FailedResources = 0
			sw.Stack.DeletedCount = 0
			sw.Stack.TotalCount = len(stackResources.StackResources)
			for _, resource := range stackResources.StackResources {
				if resource.ResourceStatus == types.ResourceStatusCreateComplete || resource.ResourceStatus == types.ResourceStatusUpdateComplete {
					sw.Stack.ReadyCount++
				}
				if resource.ResourceStatus == types.ResourceStatusCreateFailed || resource.ResourceStatus == types.ResourceStatusUpdateFailed || resource.ResourceStatus == types.ResourceStatusDeleteFailed {
					sw.Stack.FailedResources++
				}
				if resource.ResourceStatus == types.ResourceStatusDeleteComplete {
					sw.Stack.DeletedCount++
				}
			}
			if sw.Stack.ReadyCount < sw.Stack.TotalCount || sw.Stack.FailedResources > 0 {
				sw.Stack.Transitioning = true
			}

			if sw.Stack.ready() && sw.Stack.Transitioning {
				sw.Stack.Transitioning = false
				sw.emit()
				continue
			} else if sw.Stack.ready() {
				continue
			}

			sw.emit()
			time.Sleep(30 * time.Second)
		}
	}
}

func (s *Stack) ready() bool {
	return s.TotalCount == s.ReadyCount
}

func (sw *StackWatcher) emit() {
	e := &apiv1.Event{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Event",
			APIVersion: "api.acorn.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    os.Getenv(acornProjectKey),
			GenerateName: "se-",
		},

		Type:        getEventPhrase(sw.Stack.Transitioning),
		AppName:     os.Getenv(acornNameKey),
		Severity:    "info",
		Description: sw.Stack.message(),
		Resource: &internalv1.EventResource{
			Kind: "app",
			Name: os.Getenv(acornNameKey),
		},
	}
	if err := sw.KClient.Create(sw.Context, e); err != nil {
		logrus.Error(err)
	}

	fmt.Print(sw.Stack.message())
}

func getEventPhrase(t bool) string {
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

func (s *Stack) message() string {
	action := "Provisioning"
	suffix := ""
	if s.FailedResources > 0 {
		suffix = fmt.Sprintf(" (%d failed, stack: %s)", s.FailedResources, s.Name)
	}

	value := s.ReadyCount
	if os.Getenv(acornEventKey) == "delete" {
		action = "Deleting"
		value = s.DeletedCount
	}
	return fmt.Sprintf("CloudFormation: %s resources (%d/%d ready)%s\n", action, value, s.TotalCount, suffix)
}
