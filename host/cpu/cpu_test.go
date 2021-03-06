// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package cpu

import (
	"testing"
	"time"
)

func TestMaxSpeed(t *testing.T) {
	s := MaxSpeed()
	if isLinux {
		// MaxSpeed() is 0 when running in a docker container, i.e. travis-ci.org.
	} else {
		if s != 0 {
			t.Fatal("MaxSpeed() is not supported on non-linux")
		}
	}
}

func TestNanospin(t *testing.T) {
	Nanospin(time.Microsecond)
}
