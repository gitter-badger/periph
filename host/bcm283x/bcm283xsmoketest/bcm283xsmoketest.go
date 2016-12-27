// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// bcm283xsmoketest verifies that bcm283x specific functionality work.
//
// This test assumes GPIO6 and GPIO13 are connected together. GPIO6 implements
// GPCLK2 and GPIO13 imlements PWM1_OUT.
package bcm283xsmoketest

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host/bcm283x"
)

type SmokeTest struct {
	// start is to display the delta in µs.
	start time.Time
}

func (s *SmokeTest) Name() string {
	return "bcm283x"
}

func (s *SmokeTest) Description() string {
	return "Tests advanced bcm283x functionality"
}

func (s *SmokeTest) Run(args []string) error {
	if !bcm283x.Present() {
		return errors.New("this smoke test can only be used on a bcm283x based host")
	}
	f := flag.NewFlagSet("bcm283x", flag.ExitOnError)
	f.Parse(args)
	if f.NArg() != 0 {
		return errors.New("unsupported flags")
	}

	start := time.Now()
	pClk := &loggingPin{bcm283x.GPIO6, start}
	pPWM := &loggingPin{bcm283x.GPIO13, start}
	// First make sure they are connected together.
	if err := ensureConnectivity(pClk, pPWM); err != nil {
		return err
	}
	// Confirmed they are connected. Now ready to test.
	if err := s.testClock(pClk, pPWM); err != nil {
		return err
	}
	if err := s.testPWM(pPWM, pClk); err != nil {
		return err
	}
	return s.testDMA(pPWM, pClk)
}

// waitForEdge returns a channel that will return one bool, true if a edge was
// detected, false otherwise.
func (s *SmokeTest) waitForEdge(p gpio.PinIO) <-chan bool {
	c := make(chan bool)
	// A timeout inherently makes this test flaky but there's a inherent
	// assumption that the CPU edge trigger wakes up this process within a
	// reasonable amount of time; in term of latency.
	go func() {
		b := p.WaitForEdge(time.Second)
		// Author note: the test intentionally doesn't call p.Read() to test that
		// reading is not necessary.
		fmt.Printf("    %s -> WaitForEdge(%s) -> %t\n", since(s.start), p, b)
		c <- b
	}()
	return c
}

// testClock tests .PWM() for a clock pin.
func (s *SmokeTest) testClock(p1, p2 *loggingPin) error {
	fmt.Printf("- Testing clock\n")
	const period = 200 * time.Microsecond
	if err := p2.In(gpio.PullDown, gpio.BothEdges); err != nil {
		return err
	}
	time.Sleep(time.Microsecond)

	if err := p1.PWM(0, period); err != nil {
		return err
	}
	time.Sleep(time.Microsecond)
	if p2.Read() != gpio.Low {
		return fmt.Errorf("unexpected %s value; expected Low", p1)
	}

	if err := p1.PWM(gpio.DutyMax, period); err != nil {
		return err
	}
	time.Sleep(time.Microsecond)
	if p2.Read() != gpio.High {
		return fmt.Errorf("unexpected %s value; expected High", p1)
	}

	// A clock doesn't support arbitrary duty cycle.
	if err := p1.PWM(gpio.DutyHalf/2, period); err == nil {
		return fmt.Errorf("expected error on %s", p1)
	}

	if err := p1.PWM(gpio.DutyHalf, period); err != nil {
		return err
	}
	return nil
}

// testPWM tests .PWM() for a PWM pin.
func (s *SmokeTest) testPWM(p1, p2 *loggingPin) error {
	const period = 200 * time.Microsecond
	fmt.Printf("- Testing PWM\n")
	if err := p2.In(gpio.PullDown, gpio.BothEdges); err != nil {
		return err
	}
	time.Sleep(time.Microsecond)

	if err := p1.PWM(0, period); err != nil {
		return err
	}
	time.Sleep(time.Microsecond)
	if p2.Read() != gpio.Low {
		return fmt.Errorf("unexpected %s value; expected Low", p1)
	}

	if err := p1.PWM(gpio.DutyMax, period); err != nil {
		return err
	}
	time.Sleep(time.Microsecond)
	if p2.Read() != gpio.High {
		return fmt.Errorf("unexpected %s value; expected High", p1)
	}

	// A real PWM supports arbitrary duty cycle.
	if err := p1.PWM(gpio.DutyHalf/2, period); err != nil {
		return err
	}

	if err := p2.PWM(gpio.DutyHalf, period); err != nil {
		return err
	}
	return nil
}

// testDMA tests PinStreamReader.
func (s *SmokeTest) testDMA(p1, p2 *loggingPin) error {
	const period = 200 * time.Microsecond
	fmt.Printf("- Testing StreamRead\n")
	if err := p2.PWM(gpio.DutyHalf, period); err != nil {
		return err
	}
	// Gather 0.1 second of readings at 10kHz sampling rate.
	// TODO(maruel): Support >64kb buffer.
	b := make(gpio.Bits, 1000)
	if err := p1.ReadStream(gpio.PullDown, period/2, b); err != nil {
		return err
	}

	// Do debug the trace, uncomment the following line:
	//fmt.Printf("%s\n", hex.EncodeToString(b))

	// Sum the bits, it should be close to 50%.
	v := 0
	for _, x := range b {
		for j := 0; j < 8; j++ {
			v += int((x >> uint(j)) & 1)
		}
	}
	fraction := (100 * v) / (8 * len(b))
	if fraction < 45 || fraction > 55 {
		return fmt.Errorf("reading clock lead to %d%% bits On, expected 50%", fraction)
	}
	// TODO(maruel): There should be 10 streaks.
	return nil
}

//

func printPin(p gpio.PinIO) {
	fmt.Printf("- %s: %s", p, p.Function())
	if r, ok := p.(gpio.RealPin); ok {
		fmt.Printf("  alias for %s", r.Real())
	}
	fmt.Print("\n")
}

// since returns time in µs since the test start.
func since(start time.Time) string {
	µs := (time.Since(start) + time.Microsecond/2) / time.Microsecond
	ms := µs / 1000
	µs %= 1000
	return fmt.Sprintf("%3d.%03dms", ms, µs)
}

// loggingPin logs when its state changes.
type loggingPin struct {
	*bcm283x.Pin
	start time.Time
}

func (p *loggingPin) In(pull gpio.Pull, edge gpio.Edge) error {
	fmt.Printf("  %s %s.In(%s, %s)\n", since(p.start), p, pull, edge)
	return p.Pin.In(pull, edge)
}

func (p *loggingPin) Out(l gpio.Level) error {
	fmt.Printf("  %s %s.Out(%s)\n", since(p.start), p, l)
	return p.Pin.Out(l)
}

func (p *loggingPin) PWM(duty gpio.Duty, period time.Duration) error {
	fmt.Printf("  %s %s.PWM(%s, %s)\n", since(p.start), p, duty, period)
	return p.Pin.PWM(duty, period)
}

// ensureConnectivity makes sure they are connected together.
func ensureConnectivity(p1, p2 *loggingPin) error {
	if err := p1.In(gpio.PullDown, gpio.NoEdge); err != nil {
		return err
	}
	if err := p2.In(gpio.PullDown, gpio.NoEdge); err != nil {
		return err
	}
	time.Sleep(time.Microsecond)
	if p1.Read() != gpio.Low {
		return fmt.Errorf("unexpected %s value; expected low", p1)
	}
	if p2.Read() != gpio.Low {
		return fmt.Errorf("unexpected %s value; expected low", p2)
	}
	if err := p2.In(gpio.PullUp, gpio.NoEdge); err != nil {
		return err
	}
	time.Sleep(time.Microsecond)
	if p1.Read() != gpio.High {
		return fmt.Errorf("unexpected %s value; expected high", p1)
	}
	return nil
}