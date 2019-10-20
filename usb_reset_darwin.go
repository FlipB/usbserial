package usbserial

// +build darwin // redundant becase of filename

/*
#cgo LDFLAGS: -L . -L/usr/local/lib -framework CoreFoundation -framework IOKit
#include <IOKit/usb/IOUSBLib.h>
#include <CoreFoundation/CoreFoundation.h>

static inline CFIndex cfstring_utf8_length(CFStringRef str, CFIndex *need) {
  CFIndex n, usedBufLen;
  CFRange rng = CFRangeMake(0, CFStringGetLength(str));
  return CFStringGetBytes(str, rng, kCFStringEncodingUTF8, 0, 0, NULL, 0, need);
}


void scanit() {
	//IOServiceGetMatchingServices();
}
*/
import "C"

// inspiration:
// https://stackoverflow.com/questions/12786922/programmatically-unplug-and-replug-a-usb-device-to-load-new-driver-in-os-x#33108729
// https://github.com/boombuler/hid/blob/master/hid_darwin.go

// reset_usb is not implemented for darwin yet
// reset_usb should make the os rediscover the usb device (similar to unplug and replug device)
func reset_usb(vid, pid int) error {

	return nil
}
