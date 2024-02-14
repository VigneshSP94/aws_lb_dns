package aws

import (
	"aws_lb_dns/src/configs"
	"errors"
	"fmt"
	"strings"

	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/route53"
)

type elbv2Interface interface {
	DescribeLoadBalancers(input *elbv2.DescribeLoadBalancersInput) (*elbv2.DescribeLoadBalancersOutput, error)
	DescribeTags(input *elbv2.DescribeTagsInput) (*elbv2.DescribeTagsOutput, error)
}

type route53Interface interface {
	ListHostedZonesByName(input *route53.ListHostedZonesByNameInput) (*route53.ListHostedZonesByNameOutput, error)
	ListResourceRecordSets(input *route53.ListResourceRecordSetsInput) (*route53.ListResourceRecordSetsOutput, error)
	ChangeResourceRecordSets(input *route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error)
}

type AWSService struct {
	elbv2   elbv2Interface
	route53 route53Interface
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
	return result.LoadBalancers, nil
}

func (a *AWSService) GetZoneID(zoneName string) (string, error) {
	input := &route53.ListHostedZonesByNameInput{
		DNSName: aws.String(zoneName),
	}
	result, err := a.route53.ListHostedZonesByName(input)
	if err != nil {
		log.Printf("Error getting zone ID for %s: %s", zoneName, err.Error())
		return "", err
	}
	for _, zone := range result.HostedZones {
		if *zone.Name == zoneName {
			return *zone.Id, nil
		}
	}
	return "", errors.New(fmt.Sprintf("Zone %s not found", zoneName))
}

func (a *AWSService) GetRRSets(zoneID string) ([]*route53.ResourceRecordSet, error) {
	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
	}
	result, err := a.route53.ListResourceRecordSets(input)
	if err != nil {
		return nil, err
	}
	return result.ResourceRecordSets, nil
}

func GetAliasTargets(rrsets []*route53.ResourceRecordSet) []configs.Alias {
	aliasTargets := []configs.Alias{}
	for _, rrset := range rrsets {
		if *rrset.Type == "CNAME" && strings.Contains(*rrset.Name, "autolb-") {
			aliasTargets = append(aliasTargets, configs.Alias{CNAME: *rrset.Name, Alias: *rrset.ResourceRecords[0].Value})
		}
	}
	return aliasTargets
}

func (a *AWSService) AddRecords(lbs []*elbv2.LoadBalancer, tag string, aliases []configs.Alias, zoneID string, zone string, region string) error {
	for _, lb := range lbs {
		hasTag := false
		lbHasDNSEntry := false
		input := &elbv2.DescribeTagsInput{
			ResourceArns: []*string{lb.LoadBalancerArn},
		}
		result, err := a.elbv2.DescribeTags(input)
		if err != nil {
			return err
		}
		for _, tagDescription := range result.TagDescriptions {
			for _, t := range tagDescription.Tags {
				if *t.Key == tag {
					hasTag = true
					for _, alias := range aliases {
						if *lb.DNSName == alias.Alias {
							lbHasDNSEntry = true
						}
					}
				} else {
					fmt.Printf("skipiing %s\n", *lb.LoadBalancerName)
				}
			}
		}
		if hasTag && !lbHasDNSEntry {
			dnsName := "autolb-" + region + "-" + *lb.LoadBalancerName
			log.Printf("Found an LB without a DNS entry: %s", *lb.LoadBalancerName)
			log.Printf("Creating DNS record for %s with value %s", *lb.LoadBalancerName, "autolb-"+*lb.LoadBalancerName)
			a.createDNSRecord(zoneID, *lb.DNSName, dnsName, zone)
		}
	}
	return nil
}

func (a *AWSService) createDNSRecord(zoneID, name, value string, zone string) error {
	input := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(value + "." + zone),
						Type: aws.String("CNAME"),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(name),
							},
						},
						TTL: aws.Int64(300),
					},
				},
			},
		},
	}
	_, err := a.route53.ChangeResourceRecordSets(input)
	if err != nil {
		log.Printf("Error creating DNS record for %s with value %s: %s", name, value, err.Error())
		return err
	}
	log.Printf("Created a CNAME record %s pointing to the LB", name)
	return nil
}

func (a *AWSService) DeleteDNSRecord(zoneID string, alias configs.Alias) error {
	input := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("DELETE"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(alias.CNAME),
						Type: aws.String("CNAME"),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(alias.Alias),
							},
						},
						TTL: aws.Int64(300),
					},
				},
			},
		},
	}
	_, err := a.route53.ChangeResourceRecordSets(input)
	if err != nil {
		return err
	}
	return nil
}
