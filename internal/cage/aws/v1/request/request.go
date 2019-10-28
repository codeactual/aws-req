// Copyright (C) 2019 The CodeActual Go Environment Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package request

import (
	"bytes"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/signer/v4"

	"github.com/pkg/errors"
)

// Input holds request details for a  invocation.
type Input struct {
	Req   *http.Request
	Creds *credentials.Credentials
	Body  io.ReadSeeker
}

// Output holds response details from a  invocation.
type Output struct {
	// Res is included to avoid assuming which of its fields should be included individually.
	Res *http.Response

	TotalTime time.Duration
}

var awsSubdomainRe *regexp.Regexp

func init() {
	awsSubdomainRe = regexp.MustCompile(`^([a-z0-9-]+)(\.([a-z]+-[a-z]+-\d+))?\.amazonaws\.com$`)
}

// ParseHostname returns the service and region identifiers from a hostname.
//
// It performs minimmal validation, e.g. "ec200.amazonaws.com" does not emit an error.
// Instead, it returns errors only if the structure of the hostname fails basic checks, e.g. length.
//
// If the region is not included, "us-east-1" is returned.
func ParseHostname(hostname string) (service string, region string, err error) {
	if len(hostname) < 15 { // at least "x.amazonaws.com"
		return "", "", errors.Errorf("failed to parse domain [%s]", hostname)
	}

	parts := strings.Split(hostname[:len(hostname)-14], ".")

	switch len(parts) {
	case 2:
		return parts[0], parts[1], nil
	case 1:
		return parts[0], endpoints.UsEast1RegionID, nil
	default:
		return "", "", errors.Errorf("failed to parse domain [%s]", hostname)
	}
}

// Do signs the request using the input credentials, performs it with a new Client,
// and collects details like the total response time.
func Do(input *Input) (output *Output, err error) {
	output = &Output{}

	service, region, err := ParseHostname(input.Req.URL.Host)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var signErr error
	_, ok := input.Body.(io.ReadSeeker)
	if ok {
		_, signErr = v4.NewSigner(input.Creds).Sign(input.Req, input.Body, service, region, time.Now())
	} else {
		// Use explicit nil because both an uninitialized reqBody, or one assigned to nil,
		// lead to seg fault in aws/signer/v4/v4.go due to use as a non-nil interface value.
		_, signErr = v4.NewSigner(input.Creds).Sign(input.Req, nil, service, region, time.Now())
	}
	if signErr != nil {
		return nil, errors.Wrap(signErr, "failed to sign request")
	}

	var doErr error
	client := &http.Client{}
	reqStart := time.Now()
	output.Res, doErr = client.Do(input.Req)
	if doErr != nil {
		return nil, errors.Wrap(doErr, "failed to perform request")
	}

	output.TotalTime = time.Since(reqStart)

	return output, nil
}

func NewBodyString(s string) *bytes.Reader {
	return bytes.NewReader([]byte(s))
}
