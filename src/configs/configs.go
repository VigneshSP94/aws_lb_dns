package configs

import "time"

type Options struct {
	Zone     string
	Interval time.Duration
	Region   string
	Tag      string
}

type Alias struct {
	CNAME string
	Alias string
}
