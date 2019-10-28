// Copyright (C) 2019 The CodeActual Go Environment Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package request_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	cage_request "github.com/codeactual/aws-req/internal/cage/aws/v1/request"
)

func TestParseHostname(t *testing.T) {
	expectNoErr := []struct{ host, service, region string }{
		{host: "ec2.us-west-2.amazonaws.com", service: "ec2", region: "us-west-2"},
		{host: "iam.amazonaws.com", service: "iam", region: "us-east-1"},
	}
	expectErr := []string{
		"amazonaws.com",
		"localhost",
	}

	for _, item := range expectNoErr {
		service, region, err := cage_request.ParseHostname(item.host)
		require.NoError(t, err, item.host)
		require.Exactly(t, item.service, service, item.host)
		require.Exactly(t, item.region, region, item.host)
	}

	for _, host := range expectErr {
		service, region, err := cage_request.ParseHostname(host)
		require.Exactly(t, "", service, host)
		require.Exactly(t, "", region, host)
		require.Error(t, err, host)
	}
}
