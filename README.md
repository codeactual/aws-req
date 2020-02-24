# aws-req [![GoDoc](https://godoc.org/github.com/codeactual/aws-req?status.svg)](https://pkg.go.dev/mod/github.com/codeactual/aws-req) [![Go Report Card](https://goreportcard.com/badge/github.com/codeactual/aws-req)](https://goreportcard.com/report/github.com/codeactual/aws-req) [![Build Status](https://travis-ci.org/codeactual/aws-req.png)](https://travis-ci.org/codeactual/aws-req)

aws-req is a program which reads IAM credentials from standard environment variables to perform signed HTTPS requests to arbitrary AWS service URLs.

## Use Case

It was made as a service-agnostic version of the [test-invoke-method](https://docs.aws.amazon.com/cli/latest/reference/apigateway/test-invoke-method.html) command in the official AWS CLI.

# Usage

> To install: `go get -v github.com/codeactual/aws-req/cmd/aws-req`

## Configure

aws-req uses the standard environment variables:

- `AWS_ACCESS_KEY` -OR- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_KEY `-OR- `AWS_SECRET_ACCESS_KEY`
- `AWS_SESSION_TOKEN`

## Examples

> Display help:

```bash
aws-req --help
```

> EC2 API GET request:

```bash
aws-req https://ec2.amazonaws.com/?Action=DescribeAvailabilityZones&Version=2016-11-15
```

> API Gateway POST request:

```bash
aws-req --method POST --body='{"key":"val"}' https://X.execute-api.us-east-1.amazonaws.com/prod/endpoint
```

> API Gateway GET request w/ additional headers:

```bash
aws-req --header='{"key":"val"}' https://X.execute-api.us-east-1.amazonaws.com/prod/endpoint
```

> Run aws-req via [aws-exec-cmd](https://github.com/codeactual/aws-exec-cmd) to populate the environment with credentials from an EC2 instance role:

```bash
aws-exec-cmd role --chain instance -- aws-req --verbose https://ec2.amazonaws.com/?Action=DescribeAvailabilityZones&Version=2016-11-15
```

# Development

## Travis CI

### Config

- Generate AWS API credentials which will be added to the config file as encrypted environment variables.
  - [travis-iam.json](testdata/iam/travis-iam.json) is the IAM policy which creates the required grants. The [Travis CI IP addresses](https://docs.travis-ci.com/user/ip-addresses/) in the policy conditions may be out-of-date.
    - The IAM JSON lists the IPs in this order: `nat.gce-us-central1.travisci.net`, `nat.gce-us-east1.travisci.net`
- To configure the environment variables used by the functional test against the EC2 API, use the [Travis CLI](https://docs.travis-ci.com/user/encryption-keys/#usage) to generate the `secure` string value.
  - Each `env` item expects all key/value pairs as one string, and multiple items define multiple build permutations so that all pair sets are tested. Input an entire set, e.g. `AWS_ACCESS_KEY_ID=... AWS_SECRET_ACCESS_KEY=...`, in the `encrypt` command.
  - Launch `travis` in interactive mode `-i` and input the pair set without trailing newline.

# Development

## License

[Mozilla Public License Version 2.0](https://www.mozilla.org/en-US/MPL/2.0/) ([About](https://www.mozilla.org/en-US/MPL/), [FAQ](https://www.mozilla.org/en-US/MPL/2.0/FAQ/))

## Contributing

- Please feel free to submit issues, PRs, questions, and feedback.
- Although this repository consists of snapshots extracted from a private monorepo using [transplant](https://github.com/codeactual/transplant), PRs are welcome. Standard GitHub workflows are still used.
