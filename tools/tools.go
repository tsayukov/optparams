// This file is licensed under the terms of the MIT License (see LICENSE file)
// Copyright (c) 2025 Pavel Tsayukov p.tsayukov@gmail.com

//go:build tools

package tools

import (
	// Tools that are used during development.
	_ "golang.org/x/vuln/cmd/govulncheck"
)
