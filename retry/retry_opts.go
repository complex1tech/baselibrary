// Copyright 2025 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package retry

import (
	"time"

	"github.com/basecomplextech/baselibrary/logging"
)

// Options specifies the options for a retrier.
type Options struct {
	// Error is the error message.
	Error string

	// ErrorHandler handles retry errors.
	ErrorHandler ErrorHandler

	// Logger is the default logger if the error handler is not set.
	Logger logging.Logger

	// MinDelay is the min delay between retries.
	MinDelay time.Duration

	// MaxDelay is the max delay between retries.
	MaxDelay time.Duration

	// MaxRetries is the max retries, zero means unlimited.
	MaxRetries int
}

// Default returns the default options.
func Default() Options {
	return Options{
		Error:    "Failed to execute function",
		Logger:   logging.Stderr,
		MinDelay: 25 * time.Millisecond,
		MaxDelay: 1 * time.Second,
	}
}
