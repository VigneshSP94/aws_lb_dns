# AWS Load Balancer DNS

This project is used to manage DNS records for AWS Load Balancers.

## Overview

The application periodically checks the state of AWS Load Balancers and updates the DNS records in Route53 accordingly. It uses the AWS SDK for Go to interact with AWS services.

It does the following:

1. Authenticates with AWS and creates a service client.
2. Retrieves a list of all load balancers.
3. Retrieves the ID of the Route53 hosted zone specified by the command line options.
4. Retrieves a list of all resource record sets in the hosted zone.
5. Adds DNS records for the load balancers that don't already have records.
6. Cleans up DNS records for load balancers that no longer exist.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

- Go 1.16 or later
- AWS account with access to manage Load Balancers and Route53

### Installing

1. Clone the repository to your local machine.
2. Navigate to the project directory.
3. Run `go build` to compile the project.

## Usage

You can run the program with command line arguments to specify the AWS region, Route53 hosted zone, check interval, and the tag to use for the DNS record.

Here's an example:

```bash
./aws_lb_dns -region us-west-2 -zone example.com -interval 5m -tag Name