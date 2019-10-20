package usbserial

import (
	"fmt"
	"log"
	"sync"

	"github.com/google/gousb"
)

//
// Globally track opened devices. Driver will not claim devices claimed by another instance
var openedDevices map[usbDevice]*gousb.Device
var openedDevicesLock sync.Mutex

func init() {
	openedDevices = make(map[usbDevice]*gousb.Device)
}

type usbDevice struct {
	vid, pid  int
	bus, port int
	address   int
}

//
//

// Device represents a USB serial device
type Device struct {
	vid int
	pid int

	usbctx *gousb.Context
	device *gousb.Device
	config *gousb.Config

	interfaceMap map[int]*Port

	// ClaimInterfaces will claim the serial interfaces on call to Open
	ClaimInterfaces bool
	// ClaimAll will claim the entire usb device
	ClaimAll bool

	Logger Logger
}

// Logger describes interface used for logging
type Logger interface {
	Printf(format string, part ...interface{})
}

// Open the usb device that mathes VID PID.
// NOTE that devices might change their PID's and/or interface configurations depending on device's internal state.
func (d *Device) Open(Vid, Pid int) error {

	d.interfaceMap = make(map[int]*Port)

	d.usbctx = gousb.NewContext()
	d.vid = Vid
	d.pid = Pid

	deviceFound := false // chokes OpenDevices so that only a single device is opened at a time.
	devices, err := d.usbctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		if int(desc.Vendor) == Vid && int(desc.Product) == Pid {
			if deviceFound {
				return false
			}
			d := usbDevice{
				vid:     int(desc.Vendor),
				pid:     int(desc.Product),
				bus:     desc.Bus,
				port:    desc.Port,
				address: desc.Address,
			}

			openedDevicesLock.Lock()
			defer openedDevicesLock.Unlock()
			openDevice, foundOpenDevice := openedDevices[d]
			if !foundOpenDevice {
				// device is not already open - we can claim it
				deviceFound = true
				return true
			}

			// try to see if device is open still
			// it shouldnt be if OpenDevices sends us it's handle, surely? Or maybe...
			_, err := openDevice.ActiveConfigNum()
			if err != nil {
				// lets assume it's not open!
				openDevice.Close()
				deviceFound = true
				return true
			}
		}
		// dont open device!
		return false
	})

	if err != nil {
		return fmt.Errorf("error opening device VID:PID %x:%x: %v", d.vid, d.pid, err)
	}
	if len(devices) != 1 {
		return fmt.Errorf("unexpected number of matches (%d) of VID:PID %x:%x", len(devices), d.vid, d.pid)
	}
	if devices[0] == nil {
		return fmt.Errorf("device not found matching VID:PID %x:%x", d.vid, d.pid)
	}
	d.device = devices[0]

	// expecting only one config available.
	numConfigs := len(d.device.Desc.Configs)
	if numConfigs != 1 {
		return fmt.Errorf("this was unexpected - more than one config")
	}

	cfg, err := d.device.ActiveConfigNum()
	if err != nil {
		return fmt.Errorf("unable to get active configuration number: %v", err)
	}
	c, err := d.device.Config(cfg)
	if err != nil {
		return fmt.Errorf("unable to get active configuration: %v", err)
	}
	d.config = c

	//d.printInterfaces()

	err = d.device.SetAutoDetach(true)
	if err != nil {
		return fmt.Errorf("error setting detach mode: %v", err)
	}

	if d.ClaimInterfaces {

		if d.ClaimAll {
			err = d.claimInterfaces()
			if err != nil {
				return fmt.Errorf("error claiming device: %v", err)
			}
		} else {
			err = d.claimSerialInterfaces()
			if err != nil {
				return fmt.Errorf("error claiming interfaces: %v", err)
			}
		}

	}

	openedDevicesLock.Lock()
	// register device with global map to make sure it's not opened again
	ud := usbDevice{
		vid:     int(d.device.Desc.Vendor),
		pid:     int(d.device.Desc.Product),
		bus:     d.device.Desc.Bus,
		port:    d.device.Desc.Port,
		address: d.device.Desc.Address,
	}
	openedDevices[ud] = d.device
	openedDevicesLock.Unlock()

	return nil
}

// Reset the USB device
func (d *Device) Reset() error {
	err := d.closeConfig()
	if err != nil {
		log.Fatalf("unable to close config %v \n", err)
	}

	err = d.device.Reset()
	if err != nil {
		return err
	}

	// OS specific call - TODO - This should simulate the device getting unplugged and replugged
	err = reset_usb(d.vid, d.pid)
	if err != nil {
		return fmt.Errorf("error reseting usb device: %v", err)
	}
	// TODO all of this needs to be fixed up
	// wait for device?

	return fmt.Errorf("device needs to be manually unplugged")
}

// Close the usb device
func (d *Device) Close() error {
	if d.device == nil {
		// assume already closed
		return nil
	}
	err := d.closeConfig()
	if err != nil {
		return fmt.Errorf("unable to close device: %v", err)
	}
	err = d.usbctx.Close()
	if err != nil {
		return err
	}

	openedDevicesLock.Lock()
	ud := usbDevice{
		vid:     int(d.device.Desc.Vendor),
		pid:     int(d.device.Desc.Product),
		bus:     d.device.Desc.Bus,
		port:    d.device.Desc.Port,
		address: d.device.Desc.Address,
	}
	delete(openedDevices, ud)
	openedDevicesLock.Unlock()

	d.device = nil

	return nil
}

func (d *Device) closeConfig() error {
	d.closeInterfaces()
	return d.config.Close()
}

func (d *Device) errIfClosed() error {
	if d.device == nil {
		return fmt.Errorf("device not open")
	}
	return nil
}

// printInterfaces just for debugging etc
func (d *Device) printInterfaces() {
	c := d.config
	for _, intf := range c.Desc.Interfaces {
		fmt.Printf("opening interface %d: %s (Alts: %d)\n", intf.Number, intf.String(), len(intf.AltSettings))
		for _, alt := range intf.AltSettings {
			fmt.Printf("\t%d: %s (Alt: %d)\n", alt.Number, alt.String(), alt.Alternate)

			for _, ep := range alt.Endpoints {
				fmt.Printf("\t\t%d: %s", ep.Number, ep.String())
			}
		}
	}
}

func (d *Device) printf(format string, data ...interface{}) {
	if d.Logger != nil {
		d.Logger.Printf(format, data...)
	}
}

func (d *Device) claimInterfaces() error {
	d.closeInterfaces()

	for _, intf := range d.config.Desc.Interfaces {
		usbIntf, err := d.claimSerialInterface(intf.Number)
		if err != nil {
			d.printf("could not claim interface %d: %v\n", intf.Number, err)
		}
		if usbIntf != nil {
			continue
		}

		if len(intf.AltSettings) == 0 {
			return fmt.Errorf("usb interface %d has no Alt settings", intf.Number)
		}
		uIntf, err := d.config.Interface(intf.Number, intf.AltSettings[0].Alternate)
		if err != nil {
			return err
		}

		d.interfaceMap[intf.Number] = &Port{
			iface: uIntf,
		}
	}

	if len(d.interfaceMap) == 0 {
		return fmt.Errorf("no serial interfaces found")
	}

	return nil
}

func (d *Device) claimSerialInterfaces() error {

	d.closeInterfaces()

	for _, intf := range d.config.Desc.Interfaces {
		usbIntf, err := d.claimSerialInterface(intf.Number)
		if err != nil {
			d.printf("could not claim serial interface %d: %v\n", intf.Number, err)
			continue
		}
		if usbIntf == nil {
			panic("usbIntf is nil")
		}
	}

	if len(d.interfaceMap) == 0 {
		return fmt.Errorf("no serial interfaces found")
	}

	return nil
}

func (d *Device) claimSerialInterface(intfNum int) (*Port, error) {

	usbIntf, ok := d.interfaceMap[intfNum]
	if ok {
		usbIntf.Close()
		delete(d.interfaceMap, intfNum)
	}

	for _, intf := range d.config.Desc.Interfaces {
		if intf.Number != intfNum {
			continue
		}
		if len(intf.AltSettings) == 0 {
			// error - must be atleast one, surely?
			continue
		}
		foundAlt := -1
		foundInEp := -1
		foundOutEp := -1

		for _, alt := range intf.AltSettings {

			// find id's of Bulk endpoints
			for _, ep := range alt.Endpoints {
				if ep.TransferType == gousb.TransferTypeBulk {
					if ep.Direction == gousb.EndpointDirectionIn {
						if foundInEp < 0 {
							foundInEp = ep.Number
							continue
						}
						d.printf("Notice: usb interface %d Alt %d has multiple BULK IN Endpoints. Assuming %d is correct.", alt.Number, alt.Alternate, ep.Number)
					}
					if ep.Direction == gousb.EndpointDirectionOut {
						if foundOutEp < 0 {
							foundOutEp = ep.Number
							continue
						}

						d.printf("Notice: usb interface %d Alt %d has multiple BULK OUT Endpoints. Assuming %d is correct.", alt.Number, alt.Alternate, ep.Number)
					}
				}
			}

			if foundInEp < 0 || foundOutEp < 0 {
				// cannot be a serial port, need bulk in and out endpoints.
				continue
			}
			if foundAlt < 0 {
				foundAlt = alt.Alternate
				continue
			}
			d.printf("Notice: usb interface %d has multiple Alternative modes with possible serial endpoints. Assuming %d is correct one.", alt.Number, alt.Alternate)
		}

		if foundAlt < 0 {
			continue
		}

		usbIntf, err := d.config.Interface(intf.Number, foundAlt)
		if err != nil {
			return nil, err
		}

		in, err := usbIntf.InEndpoint(foundInEp)
		if err != nil {
			return nil, err
		}
		out, err := usbIntf.OutEndpoint(foundOutEp)
		if err != nil {
			return nil, err
		}

		i := &Port{
			iface: usbIntf,
			in:    in,
			out:   out,
		}
		d.interfaceMap[intf.Number] = i
		return i, nil
	}

	return nil, fmt.Errorf("no serial interface found at index %d", intfNum)
}

func (d *Device) closeInterfaces() {
	for i := range d.interfaceMap {
		d.interfaceMap[i].Close()
		delete(d.interfaceMap, i)
	}
}

// String returns the human readable identifier of the usb device
func (d *Device) String() string {
	if d.device == nil {
		return "device closed"
	}
	return d.device.String()
}

// Interface opens one of the interaces on the device
// usbInterfaceNum is the index of the usb interface to open (same as reported by lsusb for instance)
func (d *Device) Interface(usbInterfaceNum int) (*Port, error) {
	err := d.errIfClosed()
	if err != nil {
		return nil, err
	}
	return d.claimSerialInterface(usbInterfaceNum)
}

// Interfaces opens all of the detected serial looking interfaces of the device
func (d *Device) Interfaces() (map[int]*Port, error) {
	err := d.errIfClosed()
	if err != nil {
		return nil, err
	}

	err = d.claimSerialInterfaces()
	if err != nil {
		return nil, err
	}

	m := make(map[int]*Port, len(d.interfaceMap))
	for i, p := range d.interfaceMap {
		m[i] = p
	}
	return m, err
}
