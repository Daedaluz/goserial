package serial

import (
	ioctl "github.com/daedaluz/goioctl"
	"os"
	"syscall"
	"unsafe"
)

type Termios struct {
	Iflag IFlag      /* input mode flags */
	Oflag OFlag      /* output mode flags */
	Cflag CFlag      /* control mode flags */
	Lflag LFlag      /* local mode flags */
	Line  Discipline /* line discipline */
	Cc    [19]byte   /* control characters */
}

type Termios2 struct {
	Iflag  IFlag      /* input mode flags */
	Oflag  OFlag      /* output mode flags */
	Cflag  CFlag      /* control mode flags */
	Lflag  LFlag      /* local mode flags */
	Line   Discipline /* line discipline */
	Cc     [19]byte   /* control characters */
	ISpeed uint32     /* input speed */
	OSpeed uint32     /* output speed */
}

// Control characters
const (
	// VINTR
	// (003, ETX, Ctrl-C, or also 0177, DEL, rubout) Interrupt
	// character (INTR).  Send a SIGINT signal.
	// Recognized when ISIG is set, and then not passed as input
	VINTR = iota

	// VQUIT
	// (034,  FS,  Ctrl-\) Quit character (QUIT). Send SIGQUIT signal.
	// Recognized when ISIG is set, and then not passed as input.
	VQUIT

	// VERASE
	// (0177, DEL, rubout, or 010, BS, Ctrl-H, or also #) Erase character (ERASE).
	// This erases the previous not-yet-erased character,
	// but does not erase past EOF or beginning-of-line.
	// Recognized when ICANON is set, and then not passed as input.
	VERASE

	// VKILL
	// (025, NAK, Ctrl-U, or Ctrl-X, or also @) Kill character (KILL).
	// This  erases  the input since the last EOF or beginning-of-line.
	// Recognized when ICANON is set, and then not passed as input.
	VKILL

	// VEOF
	// (004, EOT, Ctrl-D) End-of-file character (EOF).  More precisely:
	// this  character  causes the pending tty buffer to be sent to the
	// waiting user program without waiting for end-of-line.  If it  is
	// the first character of the line, the read(2) in the user program
	// returns 0, which signifies end-of-file.
	// Recognized when  ICANON is set, and then not passed as input.
	VEOF

	// VTIME
	// Timeout in deciseconds for noncanonical read (TIME).
	VTIME

	// VMIN
	// Minimum number of characters for noncanonical read (MIN).
	VMIN

	// VSWTCH
	// (not in POSIX; not supported under Linux; 0, NUL) Switch character (SWTCH).
	// Used in System V to switch shells in shell layers, a predecessor to shell job control.
	VSWTCH

	// VSTART
	// (021, DC1, Ctrl-Q) Start  character  (START).
	// Restarts output stopped by the Stop character.
	// Recognized when IXON is set, and then not passed as input.
	VSTART

	// VSTOP
	// (023, DC3, Ctrl-S) Stop character  (STOP).
	// Stop  output  until Start character typed.
	// Recognized when IXON is set, and then not passed as input.
	VSTOP

	// VSUSP
	// (032, SUB, Ctrl-Z) Suspend character (SUSP).
	// Send SIGTSTP signal.
	// Recognized when ISIG is set, and then not passed as input.
	VSUSP

	// VEOL
	// (0, NUL) Additional end-of-line character (EOL).
	// Recognized when ICANON is set.
	VEOL

	// VREPRINT
	// (not in POSIX; 022, DC2, Ctrl-R) Reprint unread characters (REPRINT).
	// Recognized when ICANON and IEXTEN are set, and then not passed as input.
	VREPRINT

	// VDISCARD
	// (not in POSIX; not supported under Linux; 017, SI, Ctrl-O) Toggle: start/stop discarding pending output.
	// Recognized when IEXTEN is set, and then not passed as input.
	VDISCARD

	// VWERASE
	// (not  in  POSIX;  027,  ETB,  Ctrl-W)  Word erase (WERASE).
	// Recognized when ICANON and IEXTEN are set, and then not  passed as input.
	VWERASE

	// VLNEXT
	// (not in POSIX; 026, SYN, Ctrl-V) Literal next (LNEXT).
	// Quotes the next input character, depriving it of a possible special meaning.
	// Recognized when IEXTEN is set, and then not passed as input.
	VLNEXT

	// VEOL2
	// (not in POSIX; 0, NUL) Yet another end-of-line character (EOL2).
	// Recognized when ICANON is set.
	VEOL2
)

type IFlag uint32

// Input flags
const (
	IGNBRK  = IFlag(0000001)
	BRKINT  = IFlag(0000002)
	IGNPAR  = IFlag(0000004)
	PARMRK  = IFlag(0000010)
	INPCK   = IFlag(0000020)
	ISTRIP  = IFlag(0000040)
	INLCR   = IFlag(0000100)
	IGNCR   = IFlag(0000200)
	ICRNL   = IFlag(0000400)
	IUCLC   = IFlag(0001000)
	IXON    = IFlag(0002000)
	IXANY   = IFlag(0004000)
	IXOFF   = IFlag(0010000)
	IMAXBEL = IFlag(0020000)
	IUTF8   = IFlag(0040000)
)

type OFlag uint32

// Output flags
const (
	OPOST  = OFlag(0000001)
	OLCUC  = OFlag(0000002)
	ONLCR  = OFlag(0000004)
	OCRNL  = OFlag(0000010)
	ONOCR  = OFlag(0000020)
	ONLRET = OFlag(0000040)
	OFILL  = OFlag(0000100)
	OFDEL  = OFlag(0000200)
	NLDLY  = OFlag(0000400)
	NL0    = OFlag(0000000)
	NL1    = OFlag(0000400)
	CRDLY  = OFlag(0003000)
	CR0    = OFlag(0000000)
	CR1    = OFlag(0001000)
	CR2    = OFlag(0002000)
	CR3    = OFlag(0003000)
	TABDLY = OFlag(0014000)
	TAB0   = OFlag(0000000)
	TAB1   = OFlag(0004000)
	TAB2   = OFlag(0010000)
	TAB3   = OFlag(0014000)
	XTABS  = OFlag(0014000)
	BSDLY  = OFlag(0020000)
	BS0    = OFlag(0000000)
	BS1    = OFlag(0020000)
	VTDLY  = OFlag(0040000)
	VT0    = OFlag(0000000)
	VT1    = OFlag(0040000)
	FFDLY  = OFlag(0100000)
	FF0    = OFlag(0000000)
	FF1    = OFlag(0100000)
)

type CFlag uint32

// Control flags
const (
	CBAUD    = CFlag(0010017)
	B0       = CFlag(0000000)
	B50      = CFlag(0000001)
	B75      = CFlag(0000002)
	B110     = CFlag(0000003)
	B134     = CFlag(0000004)
	B150     = CFlag(0000005)
	B200     = CFlag(0000006)
	B300     = CFlag(0000007)
	B600     = CFlag(0000010)
	B1200    = CFlag(0000011)
	B1800    = CFlag(0000012)
	B2400    = CFlag(0000013)
	B4800    = CFlag(0000014)
	B9600    = CFlag(0000015)
	B19200   = CFlag(0000016)
	B38400   = CFlag(0000017)
	EXTA     = B19200
	EXTB     = B38400
	CSIZE    = CFlag(0000060)
	CS5      = CFlag(0000000)
	CS6      = CFlag(0000020)
	CS7      = CFlag(0000040)
	CS8      = CFlag(0000060)
	CSTOPB   = CFlag(0000100)
	CREAD    = CFlag(0000200)
	PARENB   = CFlag(0000400)
	PARODD   = CFlag(0001000)
	HUPCL    = CFlag(0002000)
	CLOCAL   = CFlag(0004000)
	CBAUDEX  = CFlag(0010000)
	BOTHER   = CFlag(0010000)
	B57600   = CFlag(0010001)
	B115200  = CFlag(0010002)
	B230400  = CFlag(0010003)
	B460800  = CFlag(0010004)
	B500000  = CFlag(0010005)
	B576000  = CFlag(0010006)
	B921600  = CFlag(0010007)
	B1000000 = CFlag(0010010)
	B1152000 = CFlag(0010011)
	B1500000 = CFlag(0010012)
	B2000000 = CFlag(0010013)
	B2500000 = CFlag(0010014)
	B3000000 = CFlag(0010015)
	B3500000 = CFlag(0010016)
	B4000000 = CFlag(0010017)
	CIBAUD   = CFlag(002003600000) /* input baud rate */
	CMSPAR   = CFlag(010000000000) /* mark or space (stick) parity */
	CRTSCTS  = CFlag(020000000000) /* flow control */
	IBSHIFT  = CFlag(16)           /* Shift from CBAUD to CIBAUD */
)

type LFlag uint32

// Line flags
const (
	ISIG    = LFlag(0000001)
	ICANON  = LFlag(0000002)
	XCASE   = LFlag(0000004)
	ECHO    = LFlag(0000010)
	ECHOE   = LFlag(0000020)
	ECHOK   = LFlag(0000040)
	ECHONL  = LFlag(0000100)
	NOFLSH  = LFlag(0000200)
	TOSTOP  = LFlag(0000400)
	ECHOCTL = LFlag(0001000)
	ECHOPRT = LFlag(0002000)
	ECHOKE  = LFlag(0004000)
	FLUSHO  = LFlag(0010000)
	PENDIN  = LFlag(0040000)
	IEXTEN  = LFlag(0100000)
	EXTPROC = LFlag(0200000)
)

type Flow uint32

const (
	TCOOFF = Flow(iota)
	TCOON
	TCIOFF
	TCION
)

type Queue uint32

const (
	TCIFLUSH = Queue(iota)
	TCOFLUSH
	TCIOFLUSH
)

type Action int

const (
	// TCSANOW
	// the change occurs immediately.
	TCSANOW = Action(iota)

	// TCSADRAIN
	// the change occurs after all output written to fd has been transmitted.
	// This option should be  used  when  changing  parameters that affect output.
	TCSADRAIN

	// TCSAFLUSH
	// the  change  occurs  after  all output written to the object
	// referred by fd has been transmitted, and all input that  has  been
	// received  but  not  read  will be discarded before the change is made
	TCSAFLUSH
)

type ModemLine int

const (
	// TIOCM_LE
	// LE / DSR (line enable / data set ready)
	TIOCM_LE = ModemLine(0x001)

	// TIOCM_DTR
	// DTR (data terminal ready)
	TIOCM_DTR = ModemLine(0x002)

	// TIOCM_RTS
	// RTS (request to send)
	TIOCM_RTS = ModemLine(0x004)

	// TIOCM_ST
	// Secondary TXD (transmit)
	TIOCM_ST = ModemLine(0x008)

	// TIOCM_SR
	// Secondary RXD (receive)
	TIOCM_SR = ModemLine(0x010)

	// TIOCM_CTS
	// CTS (clear to send)
	TIOCM_CTS = ModemLine(0x020)

	// TIOCM_CAR
	// DCD (data carrier detect)
	TIOCM_CAR = ModemLine(0x040)
	// TIOCM_CD see TIOCM_CAR
	TIOCM_CD = TIOCM_CAR

	// TIOCM_RNG
	// RNG (ring)
	TIOCM_RNG = ModemLine(0x080)
	// TIOCM_RI see TIOCM_RNG
	TIOCM_RI = TIOCM_RNG

	// TIOCM_DSR
	// DSR (data set ready)
	TIOCM_DSR = ModemLine(0x100)

	// TIOCM_OUT1
	// Unassigned programmable output 1
	TIOCM_OUT1 = ModemLine(0x2000)
	// TIOCM_OUT2
	// Unassigned programmable output 2
	TIOCM_OUT2 = ModemLine(0x4000)

	// TIOCM_LOOP
	// loopback
	TIOCM_LOOP = ModemLine(0x8000)
)

type Discipline byte

const (
	N_TTY = Discipline(iota)
	N_SLIP
	N_MOUSE
	N_PPP
	N_STRIP
	N_AX25
	N_X25
	N_6PACK
	N_MASC
	N_R3964
	N_PROFIBUS_FDL
	N_IRDA
	N_SMSBLOCK
	N_HDLC
	N_SYNC_PPP
	N_HCI
)

func Open(name string) (*Port, error) {
	f, err := os.OpenFile(name, os.O_RDWR|syscall.O_NOCTTY|syscall.SYS_SYNC, 0)
	if err != nil {
		return nil, err
	}
	return &Port{
		f: f,
	}, nil
}

type Port struct {
	f *os.File
}

func (p *Port) Write(data []byte) (n int, err error) {
	return p.f.Write(data)
}

func (p *Port) Read(data []byte) (n int, err error) {
	return p.f.Read(data)
}

func (p *Port) Close() error {
	return p.f.Close()
}

func (p *Port) GetAttr() (*Termios, error) {
	attrs := &Termios{}
	err := ioctl.Ioctl(int(p.f.Fd()), tcgets, uintptr(unsafe.Pointer(attrs)))
	if err != nil {
		return nil, err
	}
	return attrs, nil
}

func (p *Port) SetAttr(when Action, attrs *Termios) error {
	return ioctl.Ioctl(int(p.f.Fd()), tcsets+uintptr(when), uintptr(unsafe.Pointer(attrs)))
}

func (p *Port) GetAttr2() (*Termios2, error) {
	attrs := &Termios2{}
	err := ioctl.Ioctl(int(p.f.Fd()), tcgets2, uintptr(unsafe.Pointer(attrs)))
	if err != nil {
		return nil, err
	}
	return attrs, nil
}

func (p *Port) SetAttr2(when Action, attrs *Termios2) error {
	return ioctl.Ioctl(int(p.f.Fd()), tcsets2+uintptr(when), uintptr(unsafe.Pointer(attrs)))
}

// SendBreak
// If the terminal is using asynchronous serial data
// transmission, and arg is zero, then send a break (a stream
// of zero bits) for between 0.25 and 0.5 seconds.  If the
// terminal is not using asynchronous serial data
// transmission, then either a break is sent, or the function
// returns without doing anything.  When arg is nonzero,
// nobody knows what will happen.
//
// (SVr4, UnixWare, Solaris, and Linux treat
// tcsendbreak(fd,arg) with nonzero arg like tcdrain(fd).
// SunOS treats arg as a multiplier, and sends a stream of
// bits arg times as long as done for zero arg.  DG/UX and
// AIX treat arg (when nonzero) as a time interval measured
// in milliseconds.  HP-UX ignores arg.)
func (p *Port) SendBreak(arg int) error {
	return ioctl.Ioctl(int(p.f.Fd()), tcsbrk, uintptr(arg))
}

// SendBreakPosix
// So-called "POSIX version" of TCSBRK.  It treats nonzero
// arg as a time interval measured in deciseconds, and does
// nothing when the driver does not support breaks.
func (p *Port) SendBreakPosix(arg int) error {
	return ioctl.Ioctl(int(p.f.Fd()), tcsbrkp, uintptr(arg))
}

// SetBreak
// Turn break on, that is, start sending zero bits.
func (p *Port) SetBreak() error {
	return ioctl.Ioctl(int(p.f.Fd()), tiocsbrk, 1)
}

// ClearBreak
// Turn break off, that is, stop sending zero bits.
func (p *Port) ClearBreak() error {
	return ioctl.Ioctl(int(p.f.Fd()), tioccbrk, 1)
}

// Drain
// waits until all output written to the Port has been transmitted.
func (p *Port) Drain() error {
	return ioctl.Ioctl(int(p.f.Fd()), tcsbrk, 1)
}

// Flush
// discards data written to the Port but not transmitted,
// or data received but not read, depending on the queue
func (p *Port) Flush(queue Queue) error {
	return ioctl.Ioctl(int(p.f.Fd()), tcflsh, uintptr(queue))
}

// Flow
// suspends transmission or reception of data on the Port,
// depending on the flow value
func (p *Port) Flow(flow Flow) error {
	return ioctl.Ioctl(int(p.f.Fd()), tcxonc, uintptr(flow))
}

// MakeRaw
// Sets the Port to a "raw" mode
func (p *Port) MakeRaw() error {
	attrs, err := p.GetAttr()
	if err != nil {
		return err
	}
	attrs.MakeRaw()
	return p.SetAttr(TCSANOW, attrs)
}

// SetModemLines
// Set the status of modem bits.
func (p *Port) SetModemLines(line ModemLine) error {
	return ioctl.Ioctl(int(p.f.Fd()), tiocmset, uintptr(unsafe.Pointer(&line)))
}

// GetModemLines
// Get the status of modem bits.
func (p *Port) GetModemLines() (ModemLine, error) {
	var line ModemLine
	err := ioctl.Ioctl(int(p.f.Fd()), tiocmget, uintptr(unsafe.Pointer(&line)))
	return line, err
}

// EnableModemLines
// Set the indicated modem bits.
func (p *Port) EnableModemLines(line ModemLine) error {
	return ioctl.Ioctl(int(p.f.Fd()), tiocmbis, uintptr(unsafe.Pointer(&line)))
}

// DisableModemLines
// Clear the indicated modem bits.
func (p *Port) DisableModemLines(line ModemLine) error {
	return ioctl.Ioctl(int(p.f.Fd()), tiocmbic, uintptr(unsafe.Pointer(&line)))
}

func (attrs *Termios) MakeRaw() {
	attrs.Iflag &= ^(IGNBRK | BRKINT | PARMRK | ISTRIP | INLCR | IGNCR | ICRNL | IXON)
	attrs.Oflag &= ^(OPOST)
	attrs.Lflag &= ^(ECHO | ECHONL | ICANON | ISIG | IEXTEN)
	attrs.Cflag &= ^(CSIZE | PARENB)
	attrs.Cflag |= CS8
}

func (attrs *Termios2) MakeRaw() {
	attrs.Iflag &= ^(IGNBRK | BRKINT | PARMRK | ISTRIP | INLCR | IGNCR | ICRNL | IXON)
	attrs.Oflag &= ^(OPOST)
	attrs.Lflag &= ^(ECHO | ECHONL | ICANON | ISIG | IEXTEN)
	attrs.Cflag &= ^(CSIZE | PARENB)
	attrs.Cflag |= CS8
}

func (attrs *Termios) SetSpeed(speed CFlag) {
	attrs.Cflag &= ^(CBAUD)
	attrs.Cflag |= speed
}

func (attrs *Termios2) SetSpeed(speed CFlag) {
	attrs.Cflag &= ^(CBAUD)
	attrs.Cflag |= speed
}

func (attrs *Termios2) SetCustomIOSpeed(ispeed, ospeed uint32) {
	attrs.Cflag &= ^(CBAUD)
	attrs.Cflag |= BOTHER
	attrs.ISpeed = ispeed
	attrs.OSpeed = ospeed
}

func (attrs *Termios2) SetCustomSpeed(speed uint32) {
	attrs.Cflag &= ^(CBAUD)
	attrs.Cflag |= BOTHER
	attrs.ISpeed = speed
	attrs.OSpeed = speed
}
