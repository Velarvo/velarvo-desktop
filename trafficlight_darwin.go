//go:build darwin

package main

import "github.com/Velarvo/velarvo-desktop/internal/trafficlight"

func setupTrafficLights() {
	trafficlight.SetupTrafficLights(20, 10, 10, 20)
}
