package usbserial

import (
	"io"

	"github.com/google/gousb"
)

var _ io.ReadWriteCloser = &Port{}

// Port represents a serial interface on a usb device
type Port struct {
	iface *gousb.Interface
	in    *gousb.InEndpoint
	out   *gousb.OutEndpoint
}

// Close the port
func (i *Port) Close() error {

	//i._closed = true
	i.iface.Close()
	return nil
}

// String returns the full description of the port
func (i *Port) String() string {
	return i.iface.String() + " " + i.in.String() + " " + i.out.String()
}

// Read bytes from the serial device
func (i *Port) Read(buf []byte) (int, error) {
	return i.in.Read(buf)
}

// Write bytes to the serial device
func (i *Port) Write(buf []byte) (int, error) {
	return i.out.Write(buf)
}
