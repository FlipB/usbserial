package usbserial

// +build linux // redudant because of filename

import (
	"fmt"
	"syscall"
	// "github.com/vtolstov/go-ioctl"
)

// reset_usb is not implemented for linux yet
// reset_usb should make the os rediscover the usb device (similar to unplug and replug device)
func reset_usb(vid, pid int) error {

	// FIXME this is not fully implemented
	return nil

	devPath := "/dev/bus/usb/001/004"
	// TODO
	// find path of usb device (example: "/dev/bus/usb/001/004")
	// from vid and pid
	// ls /sys/bus/usb/*/idVendor
	// ls /sys/bus/usb/*/idProduct

	
	fd, _ := syscall.Open(devPath, syscall.O_RDWR, 0777)
	var USBDEVFS_RESET := 0 // ioctl.IO('U', 20) // Source of params is https://github.com/torvalds/linux/blob/master/include/uapi/linux/usbdevice_fs.h 

	var ptr *int // ioctl output to this - should not be used for this ioctl
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, USBDEVFS_RESET, ptr)
	if err != nil {
		return fmt.Errorf("unable to send USBDEVFS_RESET ioctl")
	}

	return nil
}
