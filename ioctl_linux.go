package serial

import (
	ioctl "github.com/daedaluz/goioctl"
	"unsafe"
)

var (
	tcgets  = uintptr(0x5401)
	tcsets  = uintptr(0x5402)
	tcsetsw = uintptr(0x5403)
	tcsetsf = uintptr(0x5404)

	tcgets2  = ioctl.IOR('T', 0x2A, unsafe.Sizeof(Termios2{}))
	tcsets2  = ioctl.IOW('T', 0x2B, unsafe.Sizeof(Termios2{}))
	tcsetsw2 = ioctl.IOW('T', 0x2C, unsafe.Sizeof(Termios2{}))
	tcsetsf2 = ioctl.IOW('T', 0x2D, unsafe.Sizeof(Termios2{}))

	tiocgserial = uintptr(0x541E)
	tiocsserial = uintptr(0x541F)

	tcsbrk  = uintptr(0x5409)
	tcsbrkp = uintptr(0x5425)

	tiocsbrk = uintptr(0x5427)
	tioccbrk = uintptr(0x5428)

	tcflsh = uintptr(0x540B)

	tcxonc = uintptr(0x540A)

	tiocmget = uintptr(0x5415) // get status
	tiocmbis = uintptr(0x5416) // set indicated bits
	tiocmbic = uintptr(0x5417) // clear indicated bits
	tiocmset = uintptr(0x5418) // set status

	tiocgrs485 = uintptr(0x542E)
	tiocsrs485 = uintptr(0x542F)

	tiocswinsz = uintptr(0x5414)
	tiocgwinsz = uintptr(0x5413)

	tiocpkt  = uintptr(0x5420)
	tiocgpkt = ioctl.IOR('T', 0x38, unsafe.Sizeof(int32(0)))

	tiocgptn    = ioctl.IOR('T', 0x30, unsafe.Sizeof(uint32(0)))
	tiocsptlck  = ioctl.IOW('T', 0x31, unsafe.Sizeof(int32(0)))
	tiocgptlck  = ioctl.IOR('T', 0x39, unsafe.Sizeof(int32(0)))
	tiocgptpeer = ioctl.IO('T', 0x41)

	tiocgpgrp = uintptr(0x540F)
	tiocspgrp = uintptr(0x5410)
	tiocgsid  = uintptr(0x5429)

	tiocexcl  = uintptr(0x540C)
	tiocnxcl  = uintptr(0x540D)
	tiocgexcl = uintptr(0x540E)
)
