package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/route53"
)

type AWSService struct {
	elbv2   *elbv2.ELBV2
	route53 *route53.Route53
}

func AWSAuth(region string) *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		panic(err.Error())
	}
	return sess
}

func NewAWSService(sess *session.Session) *AWSService {
	return &AWSService{
		elbv2:   elbv2.New(sess),
		route53: route53.New(sess),
	}
}

func (a *AWSService) GetLoadBalancers() ([]*elbv2.LoadBalancer, error) {
	input := &elbv2.DescribeLoadBalancersInput{}
	result, err := a.elbv2.DescribeLoadBalancers(input)
	if err != nil {
		return nil, err
	}
	for _, lb := range result.LoadBalancers {
		println(*lb.LoadBalancerName)
	}
	return result.LoadBalancers, nil
}
