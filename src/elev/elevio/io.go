package elevio

//#cgo CFLAGS:-std=c99 -g -Wall -O2 -I . -MMD
//#cgo LDFLAGS: -lpthread -lcomedi -g -lm
//#include <comedilib.h>
//#include "io.h"
//#include "channels.h"
import "C"
import (
	"log"
)

func intToBool(i C.int) bool {
	if i == 0 {
		return false
	} else {
		return true
	}
}

func IoInit() bool {
	initErr, err := C.io_init()

	if err != nil {
		log.Fatal("Initialization error of the C driver. Error: ", err)
	}
	return intToBool(initErr)
}

func IoSetBit(channel int) {
	_, err := C.io_set_bit(C.int(channel))

	if err != nil {
		log.Fatal("Unable to set bit via C driver. Error: ", err)
	}
}

func IoClearBit(channel int) {
	_, err := C.io_clear_bit(C.int(channel))

	if err != nil {
		log.Fatal("Unable to clear bit via C driver. Error: ", err)
	}
}

func IoWriteAnalog(channel, val int) {
	_, err := C.io_write_analog(C.int(channel), C.int(val))

	if err != nil {
		log.Fatal("Unable to write analog value via C driver. Error: ", err)
	}
}

func IoReadBit(channel int) bool {
	n, err := C.io_read_bit(C.int(channel))

	if err != nil {
		log.Fatal("Unable to read bit via C driver. Error: ", err)
	}
	return intToBool(n)
}

func IoReadAnalog(channel int) int {
	n, err := C.io_read_analog(C.int(channel))

	if err != nil {
		log.Fatal("Unable to read analog value via C driver. Error: ", err)
	}
	return int(n)
}