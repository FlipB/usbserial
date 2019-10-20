package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/flipb/usbserial"
)

func main() {

	vid := flag.Int("vid", 0x12d1, "usb device VID (example 0x12d1)")
	pid := flag.Int("pid", 0x1001, "usb device PID (example 0x1001)")
	interfaceNum := flag.Int("interface", -1, "interface id to open")
	usePrefix := flag.Bool("useprefix", false, "prefix lines output to stdout with interface id")
	readAll := flag.Bool("readall", false, "read from all interfaces and pipe output to stdout. Lines are prefixed by interface id followed by the '>'-characted.")

	flag.Parse()

	d := &usbserial.Device{}
	err := d.Open(*vid, *pid)
	if err != nil {
		log.Fatalf("error attaching driver to device %x:%x: %v", *vid, *pid, err)
	}
	defer d.Close()

	ports, err := d.Interfaces()
	if err != nil {
		log.Fatalf("error opening interfaces on device %x:%x: %v", *vid, *pid, err)
	}

	if len(ports) == 0 {
		log.Fatalf("no ports found")
	}

	if *interfaceNum == -1 {

		fmt.Print("interface missing. Use -interface flag to pass the id of one of the following interaces:\n")
		for i, p := range ports {
			fmt.Printf("\t%d : %s\n", i, p.String())
		}
		fmt.Print("\n")
		return
	}

	if *readAll == true {
		// forward all ports output to stdout
		for i, p := range ports {
			if *interfaceNum == i {
				continue // skip interfaceNum - it's handled below
			}
			defer p.Close()
			toStdout(p, prefix(*usePrefix, i))
		}
	}

	if len(ports) < *interfaceNum {
		log.Fatalf("invalid port specified: %d", *interfaceNum)
	}
	cmdPort := ports[*interfaceNum]
	defer cmdPort.Close()

	log.Printf("%s started.\nSerial connected to interface %d. Type commands to send to device, finish with linebreaks.\nType 'EXIT' to terminate.\n", filepath.Base(os.Args[0]), *interfaceNum)

	toStdout(cmdPort, prefix(*usePrefix, *interfaceNum))

	ch := make(chan struct{}, 0)
	fromStdin(cmdPort, ch)

	<-ch
	log.Printf("Terminating.\n")
}

func prefix(usePrefix bool, interfaceNum int) string {
	if !usePrefix {
		return ""
	}

	return fmt.Sprintf("%d> ", interfaceNum)
}

func toStdout(r io.Reader, prefix string) {
	go func() {

		buf := make([]byte, 1024*4)
		for {
			//buf.Reset()
			n, err := r.Read(buf)
			if err != nil {
				log.Printf("error while reading: %v", err)
				return
			}
			if n > 0 {
				os.Stdout.WriteString(fmt.Sprintf("%s%s\n", prefix, string(buf[0:n])))
			}
		}
	}()
}

func fromStdin(w io.Writer, c chan struct{}) {
	reader := bufio.NewReader(os.Stdin)
	go func() {
		defer close(c)
		for {
			text, _ := reader.ReadString('\n')
			if strings.ToUpper(text) == "EXIT\n" {
				break
			}
			if !strings.HasSuffix(text, "\r\n") && strings.HasSuffix(text, "\n") {
				text = strings.TrimSuffix(text, "\n") + "\r\n"
			}
			_, err := w.Write([]byte(text))
			if err != nil {
				log.Printf("error writing: %v", err)
				continue
			}
		}
	}()
}
