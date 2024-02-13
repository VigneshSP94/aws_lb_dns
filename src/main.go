package main

import (
	"aws_lb_dns/src/aws"
	"aws_lb_dns/src/configs"
	"flag"
	"time"
)

func main() {
	opts, err := parseOpts()
	if err != nil {
		panic(err.Error())
	}

	for {
		run(opts)
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

func run(opts configs.Options) {
	session := aws.AWSAuth(opts.Region)
	awsService := aws.NewAWSService(session)
	loadBalancers, err := awsService.GetLoadBalancers()
	if err != nil {
		panic(err.Error())
	}
	rrSets, err := awsService.GetRRSets(opts.Zone)
	if err != nil {
		panic(err.Error())
	}
	aliases := aws.GetAliasTargets(rrSets)
	zoneID := awsService.GetZoneID(opts.Zone)
	awsService.AddRecords(loadBalancers, opts.Tag, aliases, zoneID)
}
