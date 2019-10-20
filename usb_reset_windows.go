package usbserial

// discussion on how to do this in windows
// https://stackoverflow.com/questions/138394/how-to-programmatically-unplug-replug-an-arbitrary-usb-device

// reset_usb is not implemented for windows yet
// reset_usb should make the os rediscover the usb device (similar to unplug and replug device)
func reset_usb(vid, pid int) error {

	return nil
}
