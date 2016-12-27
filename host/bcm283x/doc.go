// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package bcm283x exposes the BCM283x GPIO functionality.
//
// This driver implements memory-mapped GPIO pin manipulation and leverages
// sysfs-gpio for edge detection.
//
// If you are looking at the actual implementation, open doc.go for further
// implementation details.
//
// Datasheet
//
// https://www.raspberrypi.org/wp-content/uploads/2012/02/BCM2835-ARM-Peripherals.pdf
//
// Its crowd-sourced errata: http://elinux.org/BCM2835_datasheet_errata
//
// BCM2836:
// https://www.raspberrypi.org/documentation/hardware/raspberrypi/bcm2836/QA7_rev3.4.pdf
//
// Another doc about PCM and PWM:
// https://scribd.com/doc/127599939/BCM2835-Audio-clocks
package bcm283x

// Other implementation details
//
// Upstream driver:
// https://github.com/torvalds/linux/blob/master/drivers/dma/bcm2835-dma.c
// https://github.com/torvalds/linux/blob/master/drivers/gpio
//
// Raspbian kernel:
// https://github.com/raspberrypi/linux/blob/rpi-4.4.y/drivers/dma
// https://github.com/raspberrypi/linux/blob/rpi-4.4.y/drivers/gpio
