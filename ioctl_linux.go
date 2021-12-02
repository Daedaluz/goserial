package serial

import (
	ioctl "github.com/daedaluz/goioctl"
	"unsafe"
)

var (
	tcsets2  = ioctl.IOW('T', 0x2B, unsafe.Sizeof(Termios{}))
	tcsetsw2 = ioctl.IOW('T', 0x2C, unsafe.Sizeof(Termios{}))
	tcsetsf2 = ioctl.IOW('T', 0x2D, unsafe.Sizeof(Termios{}))

	tcgets2 = ioctl.IOR('T', 0x2A, unsafe.Sizeof(Termios{}))

	tcsbrk  = uintptr(0x5409)
	tcsbrkp = uintptr(0x5425)

	tiocsbrk = uintptr(0x5427)
	tioccbrk = uintptr(0x5428)

	tcflsh = uintptr(0x540B)

	tcxonc = uintptr(0x540A)


	tiocmget = uintptr(0x5415) // get status
	tiocmbis = uintptr(0x5416) // set indicated bits
	tiocmbic = uintptr(0x6417) // clear indicated bits
	tiocmset = uintptr(0x5418) // set status
)
