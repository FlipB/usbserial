# usbserial

Open serial ports on usb devices as `io.ReadWriteCloser`s.

`usbserial` is a thin wrapper around gousb/libusb to help with opening usb serial devices.
Primarily tested on a Huawei K3520 3g dongle.

## Usage

```go

import (
	"fmt"
	"github.com/flipb/usbserial"
)

func main() {
	device := &usbserial.Device{}
	err = device.Open(ids.vid, ids.pid)
	if err == nil {
		fmt.Printf("unable to open device: %v\n", err)
		return
	}

	// open the device's first bulk interface, on the K3520 this is usally the command port
	commandPort, err := device.Interface(0)
	if err != nil {
		fmt.Printf("unable to open command port: %v\n", err)
		return
	}
	// returned port, commandPort is an io.ReadWriteCloser
	defer commandPort.Close()

	// Read some bytes
	n, err := commandPort.Read([]byte{})
	// etc...

	// If your device is an AT-speaking device, you can use github.com/flipb/at
	return
}


```

# Dependencies

gousb (github.com/google/gousb) and libusb-1.0

## cgo

gousb requires libusb-1.0 and by extension a c compiler. See github.com/google/gousb for instructions.

# TODOs

Implement usb_reset to make the operating system rediscover the device.
