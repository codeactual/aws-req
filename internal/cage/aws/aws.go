// Copyright (C) 2019 The CodeActual Go Environment Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package aws

import (
	"os"
)

// GetEnvAccessKey returns the environment key/value pair, checking first for AWS_ACCESS_KEY_ID
// then AWS_ACCESS_KEY.
func GetEnvAccessKey() (key, val string) {
	if candidate := os.Getenv("AWS_ACCESS_KEY_ID"); candidate != "" {
		return "AWS_ACCESS_KEY_ID", candidate
	}
	if candidate := os.Getenv("AWS_ACCESS_KEY"); candidate != "" {
		return "AWS_ACCESS_KEY", candidate
	}
	return "", ""
}

// GetEnvSecretAccessKey returns the environment key/value pair, checking first for AWS_ACCESS_KEY_ID
// then AWS_ACCESS_KEY.
func GetEnvSecretAccessKey() (key, val string) {
	if candidate := os.Getenv("AWS_SECRET_ACCESS_KEY"); candidate != "" {
		return "AWS_SECRET_ACCESS_KEY", candidate
	}
	if candidate := os.Getenv("AWS_SECRET_KEY"); candidate != "" {
		return "AWS_SECRET_KEY", candidate
	}
	return "", ""
}

// GetEnvSessionToken returns the environment key/value pair.
func GetEnvSessionToken() (key, val string) {
	return "AWS_SESSION_TOKEN", os.Getenv("AWS_SESSION_TOKEN")
}
