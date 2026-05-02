//go:build darwin

package trafficlight

/*
#cgo darwin LDFLAGS: -framework AppKit
void SetupTrafficLights(float size, float topPadding, float leftPadding, float spacing);
*/
import "C"

func SetupTrafficLights(size, topPadding, leftPadding, spacing float32) {
	C.SetupTrafficLights(C.float(size), C.float(topPadding), C.float(leftPadding), C.float(spacing))
}
