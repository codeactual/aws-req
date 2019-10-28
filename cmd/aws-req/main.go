// Copyright (C) 2019 The aws-req Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Command aws-req reads IAM credentials from standard environment variables to perform signed HTTPS requests to arbitrary AWS service URLs.
//
// Usage:
//
//   aws-req --help
//
// Environment variables:
//
//   AWS_ACCESS_KEY -OR- AWS_ACCESS_KEY_ID
//   AWS_SECRET_KEY -OR- AWS_SECRET_ACCESS_KEY
//   AWS_SESSION_TOKEN
//
// EC2 API GET request:
//
//   aws-req https://ec2.amazonaws.com/?Action=DescribeAvailabilityZones&Version=2016-11-15
//
// API Gateway POST request:
//
//   aws-req --method POST --body='{"key":"val"}' https://X.execute-api.us-east-1.amazonaws.com/prod/endpoint
//
// API Gateway GET request w/ additional headers:
//
//   aws-req --header='{"key":"val"}' https://X.execute-api.us-east-1.amazonaws.com/prod/endpoint
//
// Run aws-req via aws-exec-cmd to populate the environment with credentials from an EC2 instance role:
//
//   aws-exec-cmd role --chain instance -- aws-req --verbose https://ec2.amazonaws.com/?Action=DescribeAvailabilityZones&Version=2016-11-15
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	cage_aws "github.com/codeactual/aws-req/internal/cage/aws"
	cage_request "github.com/codeactual/aws-req/internal/cage/aws/v1/request"
	"github.com/codeactual/aws-req/internal/cage/cli/handler"
	handler_cobra "github.com/codeactual/aws-req/internal/cage/cli/handler/cobra"
	cage_reflect "github.com/codeactual/aws-req/internal/cage/reflect"
)

const (
	hr = "--"
)

func main() {
	err := handler_cobra.NewHandler(&Handler{}).Execute()
	if err != nil {
		panic(errors.WithStack(err))
	}
}

// Handler defines the sub-command flags and logic.
type Handler struct {
	handler.IO

	Body    string `usage:"HTTP body"`
	Header  string `usage:"Headers in JSON format"`
	JSON    bool   `usage:"Add application/json header"`
	Method  string `usage:"HTTP method [GET]"`
	Verbose bool   `usage:"Display headers, response time, etc."`
}

// Init defines the command, its environment variable prefix, etc.
//
// It implements cli/handler/cobra.Handler.
func (h *Handler) Init() handler_cobra.Init {
	return handler_cobra.Init{
		Cmd: &cobra.Command{
			Use:   "aws-req",
			Short: "Run an HTTP client with AWS credentials from the environment (e.g. via aws-cmd-exec)",
		},
		EnvPrefix: "AWS_REQ",
	}
}

// BindFlags binds the flags to Handler fields.
//
// It implements cli/handler/cobra.Handler.
func (h *Handler) BindFlags(cmd *cobra.Command) []string {
	cmd.Flags().BoolVarP(&h.JSON, "json", "j", true, cage_reflect.GetFieldTag(*h, "JSON", "usage"))
	cmd.Flags().BoolVarP(&h.Verbose, "verbose", "v", false, cage_reflect.GetFieldTag(*h, "Verbose", "usage"))
	cmd.Flags().StringVarP(&h.Body, "body", "d", "", cage_reflect.GetFieldTag(*h, "Body", "usage"))
	cmd.Flags().StringVarP(&h.Header, "header", "H", "", cage_reflect.GetFieldTag(*h, "Header", "usage"))
	cmd.Flags().StringVarP(&h.Method, "method", "X", "GET", cage_reflect.GetFieldTag(*h, "Method", "usage"))
	return []string{}
}

// Run performs the sub-command logic.
//
// It implements cli/handler/cobra.Handler.
func (h *Handler) Run(ctx context.Context, args []string) {
	u, parseErr := url.Parse(args[0])
	h.ExitOnErr(parseErr, "failed to parse URL", 1)
	urlStr := u.String()

	if u.Scheme != "https" {
		h.ExitOnErr(errors.New("URL must be HTTPS"), "validation failed", 1)
	}

	// Use a nil body argument because the Sign() operation will set it later.
	req, newReqErr := http.NewRequest(strings.ToUpper(h.Method), urlStr, nil)
	h.ExitOnErr(newReqErr, "failed to create request", 1)

	if h.Header != "" {
		header := http.Header{}
		jsonErr := json.Unmarshal([]byte(h.Header), &header)
		h.ExitOnErr(jsonErr, "failed to parse header JSON (are all values arrays?)", 1)

		for k, vs := range header {
			for _, v := range vs {
				req.Header.Add(k, v)
			}
		}
	}

	if h.JSON {
		req.Header.Add("Content-Type", "application/json")
	}

	if h.Verbose {
		accessStatus, secretStatus, tokenStatus := "<missing>", "<missing>", "<missing>"

		if key, val := cage_aws.GetEnvAccessKey(); key != "" && val != "" {
			accessStatus = key
		}
		fmt.Fprintf(h.Out(), "Access Key: %s\n", accessStatus)

		if key, val := cage_aws.GetEnvSecretAccessKey(); key != "" && val != "" {
			secretStatus = key
		}
		fmt.Fprintf(h.Out(), "Secret Access Key: %s\n", secretStatus)

		if key, val := cage_aws.GetEnvSessionToken(); key != "" && val != "" {
			tokenStatus = key
		}
		fmt.Fprintf(h.Out(), "Session Token: %s\n", tokenStatus)

		if len(req.Header) > 0 {
			dump, dumpErr := httputil.DumpRequestOut(req, false)
			h.ExitOnErr(dumpErr, "failed output request details", 1)
			fmt.Fprintln(h.Out(), string(dump))
		} else {
			fmt.Fprintln(h.Out(), "Request Headers: none")
		}
		fmt.Fprintln(h.Out(), hr)
	}

	awsReq := &cage_request.Input{
		Req:   req,
		Creds: credentials.NewEnvCredentials(),
	}
	if len(h.Body) > 0 {
		awsReq.Body = cage_request.NewBodyString(h.Body)
	}
	output, reqErr := cage_request.Do(awsReq)
	h.ExitOnErr(reqErr, "failed complete request", 1)

	// Show verbose output on non-200s because the response, ex. JSON, is likely not in any
	// format anticipated by the consumer anyway.
	if !h.Verbose && output.Res.StatusCode == 200 {
		_, copyErr := io.Copy(os.Stdout, output.Res.Body)
		h.ExitOnErr(copyErr, "failed to read response body", 1)
	} else {
		if len(output.Res.Header) > 0 {
			dump, dumpErr := httputil.DumpResponse(output.Res, true)
			h.ExitOnErr(dumpErr, "failed output response details", 1)
			fmt.Fprintln(h.Out(), string(dump))
		} else {
			fmt.Fprintln(h.Out(), "Response Headers: none")
		}
		fmt.Fprintln(h.Out(), hr)

		fmt.Fprintf(h.Out(), "Response Time: %s\n", output.TotalTime.Truncate(time.Millisecond))
		fmt.Fprintln(h.Out(), hr)
	}

	if output.Res.StatusCode != 200 {
		os.Exit(2)
	}
}

var _ handler_cobra.Handler = (*Handler)(nil)
