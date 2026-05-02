//go:build darwin

package doubleclick

/*
#cgo darwin LDFLAGS: -framework Cocoa -framework WebKit

void InjectDoubleClickMaximize(void);
*/
import "C"

func Enable() {
	C.InjectDoubleClickMaximize()
}
