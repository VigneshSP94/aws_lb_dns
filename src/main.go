package main

import (
	"aws_lb_dns/src/aws"
	"aws_lb_dns/src/configs"
	"flag"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/elbv2"
)

func main() {
	opts, err := parseOpts()
	if err != nil {
		panic(err.Error())
	}

	for {
		log.Println("Starting ...")
		run(opts)
		log.Printf("Sleeping for %s ...", opts.Interval)
		time.Sleep(opts.Interval)
	}
}

func parseOpts() (configs.Options, error) {
	region := flag.String("region", "", "Route53 hosted zone")
	zone := flag.String("zone", "", "Availability zone")
	interval := flag.String("interval", "5m", "Interval to check")
	tag := flag.String("tag", "Name", "Tag to use for DNS record")

	flag.Parse()

	intervalDuration, err := time.ParseDuration(*interval)
	if err != nil {
		return configs.Options{}, err
	}

	return configs.Options{
		Zone:     *zone,
		Interval: intervalDuration,
		Region:   *region,
		Tag:      *tag,
	}, nil
}

func cleanup(alias configs.Alias, lbs []*elbv2.LoadBalancer, session *aws.AWSService, zoneID string, zone string) {
	hasLB := false
	for _, lb := range lbs {
		if *lb.DNSName == alias.Alias {
			hasLB = true
		}
	}
	if !hasLB {
		log.Printf("No load balancer found for %s", alias.Alias)
		err := session.DeleteDNSRecord(zoneID, alias)
		if err != nil {
			log.Printf("Error deleting DNS record: %s", err.Error())
		} else {
			log.Printf("Deleted DNS record for %s", alias.CNAME)
		}
	}
}

func run(opts configs.Options) error {

	session := aws.AWSAuth(opts.Region)
	awsService := aws.NewAWSService(session)
	loadBalancers, err := awsService.GetLoadBalancers()
	if err != nil {
		log.Printf("Error getting load balancers: %s", err.Error())
		return err
	}
	zoneID, err := awsService.GetZoneID(opts.Zone)
	if err != nil {
		log.Printf("Error getting zone ID: %s", err.Error())
		return err
	}
	rrSets, err := awsService.GetRRSets(zoneID)
	if err != nil {
		log.Printf("Error getting resource record sets: %s", err.Error())
		return err
	}
	aliases := aws.GetAliasTargets(rrSets)
	err = awsService.AddRecords(loadBalancers, opts.Tag, aliases, zoneID, opts.Zone, opts.Region)
	if err != nil {
		log.Printf("Error adding records: %s", err.Error())
		return err
	}
	for _, alias := range aliases {
		if strings.HasPrefix(alias.CNAME, "autolb-"+opts.Region) {
			cleanup(alias, loadBalancers, awsService, zoneID, opts.Zone)
		}
	}
	return nil
}
