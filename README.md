# usbserial

Is a thin wrapper around libusb to help with opening usb serial devices.
Primarily tested on a Huawei K3520 3g dongle.

# Dependencies

gousb (github.com/google/gousb) and libusb-1.0

## cgo

gousb requires libusb-1.0 and by extension a c compiler. See github.com/google/gousb for instructions.

# TODOs

Implement usb_reset to make the operating system rediscover the device.
