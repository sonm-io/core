// +build windows

package input

import (
	"os"
	"syscall"
)

// LineSep is the separator for windows or unix systems
const LineSep = "\r\n"

// Magic constant from MSDN to control whether characters read are
// repeated back on the console.
//
// http://msdn.microsoft.com/en-us/library/windows/desktop/ms686033(v=vs.85).aspx
const ENABLE_ECHO_INPUT = 0x0004

// rawRead reads file with raw mode (without prompting to terminal).
//
// For this windows version of rawRead(). I referred the codes on
// hashicorp/vault/helper and cloudfoundry/cli/terminal
func (i *UI) rawRead(f *os.File) (string, error) {

	// In windows, Handle can be used to examine or modify the system resource.
	// https://msdn.microsoft.com/en-us/library/windows/desktop/ms724457(v=vs.85).aspx
	handle := syscall.Handle(f.Fd())

	resetFunc, err := makeRaw(handle)
	if err != nil {
		return "", err
	}
	defer resetFunc()

	return i.rawReadline(f)
}

func makeRaw(console syscall.Handle) (func(), error) {

	// Get old mode so that we can recover later
	var oldMode uint32
	if err := syscall.GetConsoleMode(console, &oldMode); err != nil {
		return nil, err
	}

	var newMode uint32 = uint32(int(oldMode) & ^ENABLE_ECHO_INPUT)
	if err := setConsoleMode(console, newMode); err != nil {
		return nil, err
	}

	return func() {
		setConsoleMode(console, oldMode)
	}, nil
}

func setConsoleMode(console syscall.Handle, mode uint32) error {
	// MustLoadDLL is like LoadDLL but panics if load operation fails.
	// LoadDLL loads DLL file into memory
	//
	// KERNEL32.DLL exposes to applications most of the Win32 base APIs,
	// such as memory management, input/output (I/O) operations,
	// process and thread creation, and synchronization functions.
	kernel32 := syscall.MustLoadDLL("kernel32")

	// Sets the input mode of a console's input buffer or
	// the output mode of a console screen buffer.
	// https://msdn.microsoft.com/en-us/library/windows/desktop/ms686033(v=vs.85).aspx
	proc := kernel32.MustFindProc("SetConsoleMode")

	r, _, err := proc.Call(uintptr(console), uintptr(mode))
	if r == 0 {
		return err
	}

	return nil
}
