package aws

import (
	"aws_lb_dns/src/configs"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/route53"
)

type elbv2Mock struct{}

type route53Mock struct{}

func (e *elbv2Mock) DescribeLoadBalancers(input *elbv2.DescribeLoadBalancersInput) (*elbv2.DescribeLoadBalancersOutput, error) {
	return &elbv2.DescribeLoadBalancersOutput{
		LoadBalancers: []*elbv2.LoadBalancer{
			{
				LoadBalancerName: aws.String("test-lb-1"),
				DNSName:          aws.String("test-lb-1.us-west-2.elb.amazonaws.com"),
			},
			{
				LoadBalancerName: aws.String("test-lb-2"),
				DNSName:          aws.String("test-lb-2.us-west-2.elb.amazonaws.com"),
			},
		},
	}, nil
}

func (e *elbv2Mock) DescribeTags(input *elbv2.DescribeTagsInput) (*elbv2.DescribeTagsOutput, error) {
	return &elbv2.DescribeTagsOutput{
		TagDescriptions: []*elbv2.TagDescription{},
	}, nil
}

func (r *route53Mock) ListHostedZonesByName(input *route53.ListHostedZonesByNameInput) (*route53.ListHostedZonesByNameOutput, error) {
	return &route53.ListHostedZonesByNameOutput{
		HostedZones: []*route53.HostedZone{
			{
				Id:   aws.String("test-zone-id"),
				Name: aws.String("test-zone-name"),
			},
		},
	}, nil
}

func (r *route53Mock) ListResourceRecordSets(input *route53.ListResourceRecordSetsInput) (*route53.ListResourceRecordSetsOutput, error) {
	return &route53.ListResourceRecordSetsOutput{
		ResourceRecordSets: []*route53.ResourceRecordSet{
			{
				Name: aws.String("auto-lb-us-east1-lb"),
				Type: aws.String("CNAME"),
				ResourceRecords: []*route53.ResourceRecord{
					{
						Value: aws.String("test-lb-1.us-west-2.elb.amazonaws.com"),
					},
				},
			},
		},
	}, nil
}

func (r *route53Mock) ChangeResourceRecordSets(input *route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error) {
	return &route53.ChangeResourceRecordSetsOutput{}, nil
}

func TestGetLoadBalancers(t *testing.T) {
	// Create a mock ELB client
	svc := &AWSService{
		elbv2: &elbv2Mock{},
	}

	// Call the DescribeLoadBalancers method
	loadBalancers, err := svc.GetLoadBalancers()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Check the result
	if len(loadBalancers) != 2 {
		t.Errorf("expected two load balancers, got %v", len(loadBalancers))
	}
}

func TestAddRecords(t *testing.T) {
	svc := &AWSService{
		elbv2: &elbv2Mock{},
	}

	zoneID := "test-zone-id"
	zone := "test-zone-name"
	region := "us-east-1"
	loadBalancers, err := svc.GetLoadBalancers()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	tag := "bleh"
	alias := []configs.Alias{{
		CNAME: "auto-test-lb-1",
		Alias: "test-lb-1.us-west-2.elb.amazonaws.com",
	}}
	err = svc.AddRecords(loadBalancers, tag, alias, zoneID, zone, region)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestDeleteDNSRecord(t *testing.T) {
	svc := &AWSService{
		route53: &route53Mock{},
	}

	zoneID := "test-zone-id"
	alias := configs.Alias{
		CNAME: "test-record-name",
		Alias: "test-lb-1.us-west-2.elb.amazonaws.com",
	}
	err := svc.DeleteDNSRecord(zoneID, alias)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestGetRRSets(t *testing.T) {
	svc := &AWSService{
		route53: &route53Mock{},
	}

	zoneID := "test-zone-id"
	_, err := svc.GetRRSets(zoneID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestGetZoneID(t *testing.T) {
	svc := &AWSService{
		route53: &route53Mock{},
	}

	zone := "test-zone-name"
	_, err := svc.GetZoneID(zone)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestGetAliasTargets(t *testing.T) {
	svc := &AWSService{
		route53: &route53Mock{},
	}
	zoneID := "test-zone-id"
	rrSets, err := svc.GetRRSets(zoneID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	aliases := GetAliasTargets(rrSets)
	if len(aliases) != 0 {
		t.Errorf("expected no aliases, got %v", len(aliases))
	}
}
