package aws

import (
	"fmt"

	"log"

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
	return result.LoadBalancers, nil
}

func (a *AWSService) GetZoneID(zoneName string) string {
	input := &route53.ListHostedZonesByNameInput{
		DNSName: aws.String(zoneName),
	}
	result, err := a.route53.ListHostedZonesByName(input)
	if err != nil {
		panic(err.Error())
	}
	for _, zone := range result.HostedZones {
		if *zone.Name == zoneName {
			println(*zone.Id)
			return *zone.Id
		}
	}
	panic(fmt.Sprintf("Zone %s not found", zoneName))
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

func GetAliasTargets(rrsets []*route53.ResourceRecordSet) []string {
	aliasTargets := []string{}
	for _, rrset := range rrsets {
		if *rrset.Type == "CNAME" {
			aliasTargets = append(aliasTargets, *rrset.AliasTarget.DNSName)
		}
	}
	return aliasTargets
}

func (a *AWSService) AddRecords(lbs []*elbv2.LoadBalancer, tag string, aliases []string, zoneID string) {
	for _, lb := range lbs {
		lbHasDNSEntry := false
		input := &elbv2.DescribeTagsInput{
			ResourceArns: []*string{lb.LoadBalancerArn},
		}
		result, err := a.elbv2.DescribeTags(input)
		if err != nil {
			panic(err.Error())
		}
		for _, tagDescription := range result.TagDescriptions {
			for _, t := range tagDescription.Tags {
				if *t.Key == tag {
					for _, alias := range aliases {
						if *lb.DNSName == alias {
							lbHasDNSEntry = true
						}
					}
				}
			}
		}
		if !lbHasDNSEntry {
			log.Printf("Creating DNS record for %s with value %s", *lb.DNSName, "lb-"+*lb.LoadBalancerName)
			a.createDNSRecord(zoneID, *lb.DNSName, "lb-"+*lb.LoadBalancerName)
		}
	}
}

func (a *AWSService) createDNSRecord(zoneID, name, value string) error {
	input := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(name),
						Type: aws.String("CNAME"),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(value),
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
	log.Printf("Created DNS record for %s with value %s", name, value)
	return nil
}
