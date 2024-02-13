package main

import (
	"aws_lb_dns/src/configs"
	"flag"
	"time"
)

func main() {
	opts, err := parseOpts()
	if err != nil {
		panic(err.Error())
	}

}

func parseOpts() (configs.Options, error) {
	zone := flag.String("zone", "", "Route53 hosted zone")
	az := flag.String("az", "", "Availability zone")
	interval := flag.String("interval", "5m", "Interval to check")
	tag := flag.String("tag", "Name", "Tag to use for DNS record")

	flag.Parse()

	intervalDuration, err := time.ParseDuration(*interval)
	if err != nil {
		return configs.Options{}, err
	}

	return configs.Options{
		Az:       *az,
		Interval: intervalDuration,
		Zone:     *zone,
		Tag:      *tag,
	}, nil
}

func run(opts configs.Options) {

}
