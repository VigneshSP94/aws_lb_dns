package configs

import "time"

type Options struct {
	Az       string
	Interval time.Duration
	Zone     string
	Tag      string
}
