package serial

import (
	"fmt"
	"github.com/daedaluz/fdev/poll"
	ioctl "github.com/daedaluz/goioctl"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
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

type Winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

type SerialFlags int32

const (
	asyncb_hup_notify = iota
	asyncb_fourport
	asyncb_sak
	asyncb_split_termios
	asyncb_spd_hi
	asyncb_spd_vhi
	asyncb_skip_test
	asyncb_auto_irq
	asyncb_session_lockout
	asyncb_pgrp_lockout
	asyncb_callout_nohup
	asyncb_hardpps_cd
	asyncb_spd_shi
	asyncb_low_latency
	asyncb_buggy_uart
	asyncb_autoprobe
	asyncb_magic_multiplier

	asyncb_suspended = 30
)
const (
	// AsyncHupNotify
	// Notify getty on hangups and closes on the callout port
	AsyncHupNotify = SerialFlags(1 << asyncb_hup_notify)

	// AsyncSuspended
	// Serial port is suspended
	AsyncSuspended = SerialFlags(1 << asyncb_suspended)

	// AsyncFourPort
	// Set OUT1, OUT2 per AST Fourport settings
	AsyncFourPort = SerialFlags(1 << asyncb_fourport)

	// AsyncSak
	// Secure Attention Key (Orange book)
	AsyncSak = SerialFlags(1 << asyncb_sak)

	// AsyncSplitTermios
	// [x] Separate termios for dialin/callout
	AsyncSplitTermios = SerialFlags(1 << asyncb_split_termios)

	// AsyncSPDHI
	// Use 57600 instead of 38400 bps
	AsyncSPDHI = SerialFlags(1 << asyncb_spd_hi)

	// AsyncSPDVHI
	// Use 115200 instead of 38400 bps
	AsyncSPDVHI = SerialFlags(1 << asyncb_spd_vhi)

	// AsyncSkipTest
	// Skip UART test during autoconfiguration
	AsyncSkipTest = SerialFlags(1 << asyncb_skip_test)

	// AsyncAutoIRQ
	// Do automatic IRQ during autoconfiguration
	AsyncAutoIRQ = SerialFlags(1 << asyncb_auto_irq)

	// AsyncSessionLockout
	// [x] Lock out cua opens based on session
	AsyncSessionLockout = SerialFlags(1 << asyncb_session_lockout)

	// AsyncPGRPLockout
	// [x] Lock out cua opens based on pgrp
	AsyncPGRPLockout = SerialFlags(1 << asyncb_pgrp_lockout)

	// AsyncCalloutNOHUP
	// [x] Don't do hangups for cua device
	AsyncCalloutNOHUP = SerialFlags(1 << asyncb_callout_nohup)

	// AsyncHardPPSCD
	// Call hardpps when CD goes high
	AsyncHardPPSCD = SerialFlags(1 << asyncb_hardpps_cd)

	// AsyncSPDSHI
	// Use 230400 instead of 38400 bps
	AsyncSPDSHI = SerialFlags(1 << asyncb_spd_shi)

	// AsyncLowLatency
	// Request low latency behaviour
	AsyncLowLatency = SerialFlags(1 << asyncb_low_latency)

	// AsyncBuggyUART
	// This is a buggy UART, skip some safety checks.  Note: can be dangerous!
	AsyncBuggyUART = SerialFlags(1 << asyncb_buggy_uart)

	// AsyncAutoProbe
	// [x] Port was autoprobed by PCI/PNP code
	AsyncAutoProbe = SerialFlags(1 << asyncb_autoprobe)

	// AsyncMagicMultiplier
	// Use special CLK or divisor
	AsyncMagicMultiplier = SerialFlags(1 << asyncb_magic_multiplier)

	AsyncSPDCust = AsyncSPDHI | AsyncSPDVHI
	AsyncSPDWarp = AsyncSPDHI | AsyncSPDSHI
	AsyncSPDMask = AsyncSPDHI | AsyncSPDVHI | AsyncSPDSHI
)

type Serial struct {
	Type          int32
	Line          int32
	Port          uint32
	Irq           int32
	Flags         SerialFlags
	XmitFifoSize  int32
	CustomDivisor int32
	BaudBase      int32
	CloseDelay    uint16
	IOType        byte
	ReservedChar  byte
	Hub6          int32
	ClosingWait   uint16 /* time to wait before closing */
	ClosingWait2  uint16 /* no longer used... */
	IOMemBase     uintptr
	IOMemRegShift uint16
	PortHigh      uint32
	IOMapBase     uint64 /* cookie passed into ioremap */
}

type RS485Flag uint32

const (
	// RS485Enabled
	// If enabled
	RS485Enabled = RS485Flag(1 << 0)
	// RS485RTSOnSend
	// Logical level for RTS pin when sending
	RS485RTSOnSend = RS485Flag(1 << 1)
	// RS485RTSAfterSend
	// Logical level for RTS pin after sent
	RS485RTSAfterSend = RS485Flag(1 << 2)
	// RS485RXDuringTx
	// Receive while transmitting
	RS485RXDuringTx = RS485Flag(1 << 4)
	// RS485TerminateBus
	// Enable bus termination (if supported)
	RS485TerminateBus = RS485Flag(1 << 5)
)

type RS485 struct {
	Flags              RS485Flag /* RS485 feature flags */
	DelayRTSBeforeSend uint32    /* Delay before send (milliseconds) */
	DelayRTSAfterSend  uint32    /* Delay after send (milliseconds) */
	padding            [5]uint32
}

// Control characters
const (
	// VINTR
	// (003, ETX, Ctrl-C, or also 0177, DEL, rubout) Interrupt
	// character (INTR). Send a SIGINT signal.
	// Recognized when ISIG is set, and then not passed as input
	VINTR = iota

	// VQUIT
	// (034, FS, Ctrl-\) Quit character (QUIT). Send SIGQUIT signal.
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
	// This erases the input since the last EOF or beginning-of-line.
	// Recognized when ICANON is set, and then not passed as input.
	VKILL

	// VEOF
	// (004, EOT, Ctrl-D) End-of-file character (EOF). More precisely:
	// this character causes the pending tty buffer to be sent to the
	// waiting user program without waiting for end-of-line. If it is
	// the first character of the line, the read(2) in the user program
	// returns 0, which signifies end-of-file.
	// Recognized when ICANON is set, and then not passed as input.
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
	// (021, DC1, Ctrl-Q) Start character (START).
	// Restarts output stopped by the Stop character.
	// Recognized when IXON is set, and then not passed as input.
	VSTART

	// VSTOP
	// (023, DC3, Ctrl-S) Stop character (STOP).
	// Stop output until Start character typed.
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
	// (not in POSIX; 027, ETB, Ctrl-W) Word erase (WERASE).
	// Recognized when ICANON and IEXTEN are set, and then not passed as input.
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
	// IGNBRK Ignore BREAK condition on input.
	IGNBRK = IFlag(0000001)

	// BRKINT If IGNBRK is set, a BREAK is ignored.
	// If it is not set but BRKINT is set, then a BREAK causes the input and output queues
	// to be flushed, and if the terminal is the controlling terminal of a foreground process group,
	// it will cause a SIGINT to be sent to this foreground process group.
	// When neither IGNBRK nor BRKINT are set, a BREAK reads as a null byte ('\0'),
	// except when PARMRK is set, in which case it reads as the sequence \377 \0 \0.
	BRKINT = IFlag(0000002)

	// IGNPAR Ignore framing errors and parity errors.
	IGNPAR = IFlag(0000004)

	// PARMRK If this bit is set, input bytes with parity or framing errors
	// are marked when passed to the program.
	// This bit is meaningful only when INPCK is set and IGNPAR is not set.
	// The way erroneous bytes are marked is with two preceding bytes, \377 and \0.
	// Thus, the program actually reads three bytes for one erroneous byte received from the terminal.
	// If a valid byte has the value \377, and ISTRIP (see below) is not set, the program might
	// confuse it with the prefix that marks a parity error.
	// Therefore, a valid byte \377 is passed to the program as two bytes, \377 \377 , in this case.
	PARMRK = IFlag(0000010)

	// INPCK Enable input parity checking.
	INPCK = IFlag(0000020)

	// ISTRIP Strip off eighth bit.
	ISTRIP = IFlag(0000040)

	// INLCR Translate NL to CR on input.
	INLCR = IFlag(0000100)

	// IGNCR Ignore carriage return on input.
	IGNCR = IFlag(0000200)

	// ICRNL Translate carriage return to newline on input (unless IGNCR is set).
	ICRNL = IFlag(0000400)

	// IUCLC (not in POSIX) Map uppercase characters to lowercase on input.
	IUCLC = IFlag(0001000)

	// IXON Enable XON/XOFF flow control on output.
	IXON = IFlag(0002000)

	// IXANY (XSI) Typing any character will restart stopped output.
	// (The default is to allow just the START character to restart output.)
	IXANY = IFlag(0004000)

	// IXOFF Enable XON/XOFF flow control on input.
	IXOFF = IFlag(0010000)

	// IMAXBEL (not in POSIX) Ring bell when input queue is full.
	// Linux does not implement this bit, and acts as if it is always set.
	IMAXBEL = IFlag(0020000)

	// IUTF8 (since Linux 2.6.4) (not in POSIX) Input is UTF8; this allows character-erase to be
	// correctly performed in cooked mode.
	IUTF8 = IFlag(0040000)
)

type OFlag uint32

// Output flags
const (
	// OPOST Enable implementation-defined output processing.
	OPOST = OFlag(0000001)

	// OLCUC (not in POSIX) Map lowercase characters to uppercase on output.
	OLCUC = OFlag(0000002)

	// ONLCR (XSI) Map NL to CR-NL on output.
	ONLCR = OFlag(0000004)

	// OCRNL Map CR to NL on output.
	OCRNL = OFlag(0000010)

	// ONOCR Don't output CR at column 0.
	ONOCR = OFlag(0000020)

	// ONLRET Don't output CR.
	ONLRET = OFlag(0000040)

	// OFILL Send fill characters for a delay, rather than using a timed delay.
	OFILL = OFlag(0000100)

	// OFDEL Fill character is ASCII DEL (0177). If unset, fill character is ASCII NUL ('\0').
	// (Not implemented on Linux.)
	OFDEL = OFlag(0000200)

	// NLDLY Newline delay mask. Values are NL0 and NL1.
	NLDLY = OFlag(0000400)
	NL0   = OFlag(0000000)
	NL1   = OFlag(0000400)

	// CRDLY Carriage return delay mask. Values are CR0, CR1, CR2, or CR3.
	CRDLY = OFlag(0003000)
	CR0   = OFlag(0000000)
	CR1   = OFlag(0001000)
	CR2   = OFlag(0002000)
	CR3   = OFlag(0003000)

	// TABDLY Horizontal tab delay mask. Values are TAB0, TAB1, TAB2, TAB3 (or XTABS, but see the BUGS section)
	// A value of TAB3, that is, XTABS, expands tabs to spaces (with tab stops every eight columns).
	TABDLY = OFlag(0014000)
	TAB0   = OFlag(0000000)
	TAB1   = OFlag(0004000)
	TAB2   = OFlag(0010000)
	TAB3   = OFlag(0014000)
	XTABS  = OFlag(0014000)

	// BSDLY Backspace delay mask. Values are BS0 or BS1. (Has never been implemented)
	BSDLY = OFlag(0020000)
	BS0   = OFlag(0000000)
	BS1   = OFlag(0020000)

	// VTDLY Vertical tab delay mask. Values are VT0 or VT1.
	VTDLY = OFlag(0040000)
	VT0   = OFlag(0000000)
	VT1   = OFlag(0040000)

	// FFDLY Form feed delay mask. Values are FF0 or FF1.
	FFDLY = OFlag(0100000)
	FF0   = OFlag(0000000)
	FF1   = OFlag(0100000)
)

type CFlag uint32

// Control flags
const (
	// CBAUD (not in POSIX) Baud speed mask (4+1 bits).
	CBAUD  = CFlag(0010017)
	B0     = CFlag(0000000)
	B50    = CFlag(0000001)
	B75    = CFlag(0000002)
	B110   = CFlag(0000003)
	B134   = CFlag(0000004)
	B150   = CFlag(0000005)
	B200   = CFlag(0000006)
	B300   = CFlag(0000007)
	B600   = CFlag(0000010)
	B1200  = CFlag(0000011)
	B1800  = CFlag(0000012)
	B2400  = CFlag(0000013)
	B4800  = CFlag(0000014)
	B9600  = CFlag(0000015)
	B19200 = CFlag(0000016)
	B38400 = CFlag(0000017)
	EXTA   = B19200
	EXTB   = B38400

	// CSIZE Character size mask. Values are CS5, CS6, CS7, or CS8.
	CSIZE = CFlag(0000060)
	// CS5 Character is 5 bit
	CS5 = CFlag(0000000)
	// CS6 Character is 6 bit
	CS6 = CFlag(0000020)
	// CS7 Character is 7 bit
	CS7 = CFlag(0000040)
	// CS8 Character is 8 bit
	CS8 = CFlag(0000060)

	// CSTOPB Set two stop bits, rather than one.
	CSTOPB = CFlag(0000100)

	// CREAD Enable receiver.
	CREAD = CFlag(0000200)

	// PARENB Enable parity generation on output and parity checking for input.
	PARENB = CFlag(0000400)

	// PARODD If set, then parity for input and output is odd; otherwise even parity is used.
	PARODD = CFlag(0001000)

	// HUPCL Lower modem control lines after last process closes the device (hang up).
	HUPCL = CFlag(0002000)

	// CLOCAL Ignore modem control lines.
	CLOCAL = CFlag(0004000)

	// CBAUDEX (not in POSIX) Extra baud speed mask (1 bit).
	// POSIX says that the baud speed is stored in the termios structure
	// without specifying where precisely, and provides cfgetispeed() and
	// cfsetispeed() for getting at it.
	// Some systems use bits selected by CBAUD in c_cflag, other systems
	// use separate fields, for example, sg_ispeed and sg_ospeed.
	CBAUDEX = CFlag(0010000)
	BOTHER  = CFlag(0010000)

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

	// CIBAUD (not in POSIX) Mask for input speeds.
	// The values for the CIBAUD bits are the same as the values for the CBAUD bits,
	// shifted left IBSHIFT bits.
	CIBAUD = CFlag(002003600000) /* input baud rate */

	// CMSPAR (not in POSIX) Use "stick" (mark/space) parity
	// (supported on certain serial devices): if PARODD is set, the parity bit is always 1;
	// if PARODD is not set, then the parity bit is always 0.
	CMSPAR = CFlag(010000000000) /* mark or space (stick) parity */

	// CRTSCTS (not in POSIX) Enable RTS/CTS (hardware) flow control.
	CRTSCTS = CFlag(020000000000) /* flow control */
	IBSHIFT = CFlag(16)           /* Shift from CBAUD to CIBAUD */
)

type LFlag uint32

// Line flags
const (
	// ISIG When any of the characters INTR, QUIT, SUSP, or DSUSP are received,
	// generate corresponding signal.
	ISIG = LFlag(0000001)

	// ICANON Enable canonical mode (described below).
	ICANON = LFlag(0000002)

	// XCASE (not in POSIX; not supported under Linux) If ICANON is also set,
	// terminal is uppercase only. Input is converted to lowercase,
	// except for characters preceded by \. On output, uppercase characters
	// are preceded by \ and lowercase characters are converted to uppercase.
	XCASE = LFlag(0000004)

	// ECHO Echo input characters.
	ECHO = LFlag(0000010)
	// ECHOE If ICANON is also set, the ERASE character erases the
	// preceding input character, and WERASE erases the preceding word.
	ECHOE = LFlag(0000020)

	// ECHOK If ICANON is also set, the KILL character erases the current line.
	ECHOK = LFlag(0000040)

	// ECHONL If ICANON is also set, echo the NL character even if ECHO is not set.
	ECHONL = LFlag(0000100)

	// NOFLSH Disable flushing the input and output queues when generating
	// signals for the INT, QUIT, and SUSP characters.
	NOFLSH = LFlag(0000200)

	// TOSTOP Send the SIGTTOU signal to the process group of a background
	// process which tries to write to its controlling terminal.
	TOSTOP = LFlag(0000400)

	// ECHOCTL (not in POSIX) If ECHO is also set, terminal special characters
	// other than TAB, NL, START, and STOP are echoed as ^X, where X is
	// the character with ASCII code 0x40 greater than the special
	// character. For example, character 0x08 (BS) is echoed as ^H.
	//
	ECHOCTL = LFlag(0001000)

	// ECHOPRT (not in POSIX) If ICANON and ECHO are also set, characters are
	// printed as they are being erased.
	ECHOPRT = LFlag(0002000)

	// ECHOKE (not in POSIX) If ICANON is also set, KILL is echoed by erasing
	// each character on the line, as specified by ECHOE and ECHOPRT.
	ECHOKE = LFlag(0004000)

	// FLUSHO (not in POSIX; not supported under Linux) Output is being
	// flushed. This flag is toggled by typing the DISCARD character.
	FLUSHO = LFlag(0010000)

	// PENDIN (not in POSIX; not supported under Linux) All characters in the
	// input queue are reprinted when the next character is read.
	PENDIN = LFlag(0040000)

	// IEXTEN Enable implementation-defined input processing.
	// This flag, as well as ICANON must be enabled for the special characters EOL2,
	// LNEXT, REPRINT, WERASE to be interpreted, and for the IUCLC flag
	// to be effective.
	IEXTEN = LFlag(0100000)

	// EXTPROC external processing
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
	// This option should be used when changing parameters that affect output.
	TCSADRAIN

	// TCSAFLUSH
	// the change occurs after all output written to the object
	// referred by fd has been transmitted, and all input that has been
	// received but not read will be discarded before the change is made
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

func (m ModemLine) String() string {
	flags := make([]string, 0, len(modemLineStrings))
	for i := 1; i <= int(TIOCM_LOOP); i <<= 1 {
		if int(m)&i > 0 {
			if flag, ok := modemLineStrings[ModemLine(i)]; ok {
				flags = append(flags, flag)
			} else {
				flags = append(flags, fmt.Sprintf("Unknown(%x)", i))
			}
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(flags, "|"))
}

var modemLineStrings = map[ModemLine]string{
	TIOCM_LE:   "LE",
	TIOCM_DTR:  "DTR",
	TIOCM_RTS:  "RTS",
	TIOCM_ST:   "ST",
	TIOCM_SR:   "SR",
	TIOCM_CTS:  "CTS",
	TIOCM_CAR:  "CAR",
	TIOCM_RNG:  "RNG",
	TIOCM_DSR:  "DSR",
	TIOCM_OUT1: "OUT1",
	TIOCM_OUT2: "OUT2",
	TIOCM_LOOP: "LOOP",
}

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

type PacketControl uint8

const (
	TIOCPKT_DATA = PacketControl(1 << iota)
	TIOCPKT_FLUSHREAD
	TIOCPKT_FLUSHWRITE
	TIOCPKT_STOP
	TIOCPKT_START
	TIOCPKT_NOSTOP
	TIOCPKT_DOSTOP
	TIOCPKT_IOCTL
)

// Options available options when oping serial port.
type Options struct {
	ReadTimeout time.Duration
	OpenMode    int
}

// NewOptions returns a new Options with default values.
// The default values are:
//
//	ReadTimeout: -1
//	OpenMode: syscall.O_RDWR | syscall.O_NOCTTY | syscall.SYS_SYNC
func NewOptions() *Options {
	return &Options{ReadTimeout: -1, OpenMode: syscall.O_RDWR | syscall.O_NOCTTY | syscall.SYS_SYNC}
}

// SetReadTimeout sets the read timeout for the serial port.
func (o *Options) SetReadTimeout(timeout time.Duration) *Options {
	o.ReadTimeout = timeout
	return o
}

// Port represents a serial port.
type Port struct {
	options *Options
	f       atomic.Value
}

// Open a serial port with the given name and options.
func Open(name string, opts *Options) (*Port, error) {
	if opts == nil {
		opts = NewOptions()
	}
	fd, err := syscall.Open(name, opts.OpenMode, 0)
	if err != nil {
		return nil, wrapErr("Open", err)
	}
	res := &Port{
		options: opts,
	}
	res.f.Store(fd)
	return res, nil
}

// NewPort returns a Port with the given file descriptor and options.
func NewPort(fd int, opts *Options) (*Port, error) {
	if opts == nil {
		opts = NewOptions()
	}
	res := &Port{
		options: opts,
	}
	res.f.Store(fd)
	return res, nil
}

// Write data to the serial port.
func (p *Port) Write(data []byte) (n int, err error) {
	x, err := syscall.Write(p.f.Load().(int), data)
	return x, wrapErr("Write", err)
}

func (p *Port) readTimeout(data []byte, timeout time.Duration) (int, error) {
	if err := poll.WaitInput(p.f.Load().(int), timeout); err != nil {
		return 0, wrapErr("", err)
	}
	n, err := syscall.Read(p.f.Load().(int), data)
	return n, wrapErr("ReadTimeout", err)
}

// Read data from the serial port.
func (p *Port) Read(data []byte) (n int, err error) {
	if p.options.ReadTimeout > -1 {
		return p.readTimeout(data, p.options.ReadTimeout)
	}
	n, err = syscall.Read(p.f.Load().(int), data)
	return n, wrapErr("Read", err)
}

// ReadTimeout reads data with timeout.
func (p *Port) ReadTimeout(data []byte, timeout time.Duration) (n int, err error) {
	return p.readTimeout(data, timeout)
}

// SetReadTimeout sets the read timeout for the serial port.
func (p *Port) SetReadTimeout(timeout time.Duration) {
	p.options.ReadTimeout = timeout
}

// Fd returns the file descriptor referencing the open serial port.
func (p *Port) Fd() int {
	return p.f.Load().(int)
}

// Close the serial port.
func (p *Port) Close() error {
	if x := p.f.Swap(-1); x != -1 {
		return wrapErr("Close", syscall.Close(x.(int)))
	}
	return ErrClosed
}

// GetAttr returns the current termios serial port settings.
func (p *Port) GetAttr() (*Termios, error) {
	attrs := &Termios{}
	err := ioctl.Ioctl(uintptr(p.f.Load().(int)), tcgets, uintptr(unsafe.Pointer(attrs)))
	if err != nil {
		return nil, wrapErr("GetAttr", err)
	}
	return attrs, nil
}

// SetAttr sets the termios serial port settings.
func (p *Port) SetAttr(when Action, attrs *Termios) error {
	return wrapErr("SetAttr", ioctl.Ioctl(uintptr(p.f.Load().(int)), tcsets+uintptr(when), uintptr(unsafe.Pointer(attrs))))
}

// GetAttr2 returns the current termios2 serial port settings.
func (p *Port) GetAttr2() (*Termios2, error) {
	attrs := &Termios2{}
	err := ioctl.Ioctl(uintptr(p.f.Load().(int)), tcgets2, uintptr(unsafe.Pointer(attrs)))
	if err != nil {
		return nil, wrapErr("GetAttr2", err)
	}
	return attrs, nil
}

// SetAttr2 sets the termios2 serial port settings.
func (p *Port) SetAttr2(when Action, attrs *Termios2) error {
	err := ioctl.Ioctl(uintptr(p.f.Load().(int)), tcsets2+uintptr(when), uintptr(unsafe.Pointer(attrs)))
	return wrapErr("SetAttr2", err)
}

// GetSerial returns the current serial port settings.
func (p *Port) GetSerial() (*Serial, error) {
	serial := &Serial{}
	err := ioctl.Ioctl(uintptr(p.f.Load().(int)), tiocgserial, uintptr(unsafe.Pointer(serial)))
	if err != nil {
		return nil, wrapErr("GetSerial", err)
	}
	return serial, nil
}

func (p *Port) SetSerial(s *Serial) error {
	return wrapErr("SetSerial", ioctl.Ioctl(uintptr(p.f.Load().(int)), tiocsserial, uintptr(unsafe.Pointer(s))))
}

// SendBreak
// If the terminal is using asynchronous serial data
// transmission, and arg is zero, then send a break (a stream
// of zero bits) for between 0.25 and 0.5 seconds. If the
// terminal is not using asynchronous serial data
// transmission, then either a break is sent, or the function
// returns without doing anything. When arg is nonzero,
// nobody knows what will happen.
//
// (SVr4, UnixWare, Solaris, and Linux treat
// tcsendbreak(fd,arg) with nonzero arg like tcdrain(fd).
// SunOS treats arg as a multiplier, and sends a stream of
// bits arg times as long as done for zero arg. DG/UX and
// AIX treat arg (when nonzero) as a time interval measured
// in milliseconds. HP-UX ignores arg.)
func (p *Port) SendBreak(arg int) error {
	return wrapErr("SendBreak", ioctl.Ioctl(uintptr(p.f.Load().(int)), tcsbrk, uintptr(arg)))
}

// SendBreakPosix
// So-called "POSIX version" of TCSBRK. It treats nonzero
// arg as a time interval measured in deciseconds, and does
// nothing when the driver does not support breaks.
func (p *Port) SendBreakPosix(arg int) error {
	return wrapErr("SendBreakPosix", ioctl.Ioctl(uintptr(p.f.Load().(int)), tcsbrkp, uintptr(arg)))
}

// SetBreak
// Turn break on, that is, start sending zero bits.
func (p *Port) SetBreak() error {
	return wrapErr("SetBreak", ioctl.Ioctl(uintptr(p.f.Load().(int)), tiocsbrk, 1))
}

// ClearBreak
// Turn break off, that is, stop sending zero bits.
func (p *Port) ClearBreak() error {
	return wrapErr("ClearBreak", ioctl.Ioctl(uintptr(p.f.Load().(int)), tioccbrk, 1))
}

// Drain
// waits until all output written to the Port has been transmitted.
func (p *Port) Drain() error {
	return wrapErr("Drain", ioctl.Ioctl(uintptr(p.f.Load().(int)), tcsbrk, 1))
}

// Flush
// discards data written to the Port but not transmitted,
// or data received but not read, depending on the queue
func (p *Port) Flush(queue Queue) error {
	return wrapErr("Flush", ioctl.Ioctl(uintptr(p.f.Load().(int)), tcflsh, uintptr(queue)))
}

// Flow
// suspends transmission or reception of data on the Port,
// depending on the flow value
func (p *Port) Flow(flow Flow) error {
	return wrapErr("Flow", ioctl.Ioctl(uintptr(p.f.Load().(int)), tcxonc, uintptr(flow)))
}

// GetRS485
// Returns current rs485 configuration
func (p *Port) GetRS485() (*RS485, error) {
	rs485cfg := &RS485{}
	err := ioctl.Ioctl(uintptr(p.f.Load().(int)), tiocgrs485, uintptr(unsafe.Pointer(rs485cfg)))
	if err != nil {
		return nil, wrapErr("GetRS485", err)
	}
	return rs485cfg, nil
}

// SetRS485
// Set rs485 parameters
func (p *Port) SetRS485(cfg *RS485) error {
	return wrapErr("SetRS485", ioctl.Ioctl(uintptr(p.f.Load().(int)), tiocsrs485, uintptr(unsafe.Pointer(cfg))))
}

// MakeRaw
// Sets the Port to a "raw" mode
func (p *Port) MakeRaw() error {
	attrs, err := p.GetAttr()
	if err != nil {
		return wrapErr("", err)
	}
	attrs.MakeRaw()
	return wrapErr("MakeRaw", p.SetAttr(TCSANOW, attrs))
}

// GetWinSize returns the window size of the terminal.
func (p *Port) GetWinSize() (*Winsize, error) {
	ws := &Winsize{}
	err := ioctl.Ioctl(uintptr(p.f.Load().(int)), tiocgwinsz, uintptr(unsafe.Pointer(ws)))
	if err != nil {
		return nil, wrapErr("GetWinSize", err)
	}
	return ws, nil
}

// SetWinSize sets the window size of the terminal.
func (p *Port) SetWinSize(ws *Winsize) error {
	return wrapErr("SetWinSize", ioctl.Ioctl(uintptr(p.f.Load().(int)), tiocswinsz, uintptr(unsafe.Pointer(ws))))
}

// SetModemLines
// Set the status of modem bits.
func (p *Port) SetModemLines(line ModemLine) error {
	return wrapErr("SetModemLines", ioctl.Ioctl(uintptr(p.f.Load().(int)), tiocmset, uintptr(unsafe.Pointer(&line))))
}

// GetModemLines
// Get the status of modem bits.
func (p *Port) GetModemLines() (ModemLine, error) {
	var line ModemLine
	err := ioctl.Ioctl(uintptr(p.f.Load().(int)), tiocmget, uintptr(unsafe.Pointer(&line)))
	return line, wrapErr("GetModemLines", err)
}

// EnableModemLines
// Set the indicated modem bits.
func (p *Port) EnableModemLines(line ModemLine) error {
	return wrapErr("EnableModemLines", ioctl.Ioctl(uintptr(p.f.Load().(int)), tiocmbis, uintptr(unsafe.Pointer(&line))))
}

// DisableModemLines
// Clear the indicated modem bits.
func (p *Port) DisableModemLines(line ModemLine) error {
	return wrapErr("DisableModemLines", ioctl.Ioctl(uintptr(p.f.Load().(int)), tiocmbic, uintptr(unsafe.Pointer(&line))))
}

// SetPacketMode Enable or disable packet mode.
// Can be applied to the master side of a pseudo-terminal only.
// In packet mode, each subsequent read(2) will return a packet that either contains a single nonzero PacketControl byte
// Or has a single byte containing zero ('\0') followed by data written on the slave side of the pseudo-terminal.
// While packet mode is in use, the presence of control status information to be read
// from the master side may be detected by select(2) for exceptional conditions or a poll(2) for the POLLPRI event.
// This mode is used by rlogin(1) and rlogind(8) to implement a remote-echoed, locally ^S/^Q flow-controlled remote login.
func (p *Port) SetPacketMode(enable bool) error {
	x := uint32(0)
	if enable {
		x = 1
	}
	return wrapErr("SetPacketMode", ioctl.Ioctl(uintptr(p.f.Load().(int)), tiocpkt, uintptr(unsafe.Pointer(&x))))
}

// GetPacketMode returns true if the terminal is in packet mode.
func (p *Port) GetPacketMode() (bool, error) {
	x := uint32(0)
	err := ioctl.Ioctl(uintptr(p.f.Load().(int)), tiocgpkt, uintptr(unsafe.Pointer(&x)))
	if err != nil {
		return false, wrapErr("GetPacketMode", err)
	}
	return x != 0, nil
}

// SetLockPT sets or remove the lock on the pseudo-terminal slave device.
func (p *Port) SetLockPT(lock bool) error {
	x := uint32(0)
	if lock {
		x = 1
	}
	return wrapErr("SetLockPT", ioctl.Ioctl(uintptr(p.f.Load().(int)), tiocsptlck, uintptr(unsafe.Pointer(&x))))
}

// GetLockPT returns true if the pseudo-terminal slave device is locked.
func (p *Port) GetLockPT() (bool, error) {
	x := uint32(0)
	err := ioctl.Ioctl(uintptr(p.f.Load().(int)), tiocgptlck, uintptr(unsafe.Pointer(&x)))
	if err != nil {
		return false, wrapErr("GetLockPT", err)
	}
	return x != 0, nil
}

// GetPTPeer Given a file descriptor referring to a pseudo-terminal master,
// returns a new Port with the descriptor of the corresponding pseudo-terminal slave.
func (p *Port) GetPTPeer(openFlags int) (*Port, error) {
	fd, err := ioctl.IoctlX(uintptr(p.f.Load().(int)), tiocgptpeer, uintptr(openFlags))
	if err != nil {
		return nil, wrapErr("GetPTPeer", err)
	}
	return NewPort(int(fd), nil)
}

// PTSNumber returns the pts number of the slave pseudo-terminal device.
func (p *Port) PTSNumber() (int, error) {
	ptn := uint32(0)
	err := ioctl.Ioctl(uintptr(p.f.Load().(int)), tiocgptn, uintptr(unsafe.Pointer(&ptn)))
	if err != nil {
		return 0, wrapErr("PTSNumber", err)
	}
	return int(ptn), nil
}

// GetGroupID returns the group ID of the slave pseudo-terminal device.
func (p *Port) GetGroupID() (int, error) {
	gid := 0
	err := ioctl.Ioctl(uintptr(p.f.Load().(int)), tiocgpgrp, uintptr(unsafe.Pointer(&gid)))
	if err != nil {
		return 0, wrapErr("GetGroupID", err)
	}
	return gid, nil
}

// SetGroupID sets the group ID of the slave pseudo-terminal device.
func (p *Port) SetGroupID(gid int) error {
	return wrapErr("SetGroupID", ioctl.Ioctl(uintptr(p.f.Load().(int)), tiocspgrp, uintptr(unsafe.Pointer(&gid))))
}

// GetSessionID returns the session ID of the slave pseudo-terminal device.
func (p *Port) GetSessionID() (int, error) {
	sid := 0
	err := ioctl.Ioctl(uintptr(p.f.Load().(int)), tiocgsid, uintptr(unsafe.Pointer(&sid)))
	if err != nil {
		return 0, wrapErr("GetSessionID", err)
	}
	return sid, nil
}

// EnableExclusiveMode enables exclusive mode for the slave pseudo-terminal device.
func (p *Port) EnableExclusiveMode() error {
	return wrapErr("EnableExclusiveMode", ioctl.Ioctl(uintptr(p.f.Load().(int)), tiocexcl, 0))
}

// DisableExclusiveMode disables exclusive mode for the slave pseudo-terminal device.
func (p *Port) DisableExclusiveMode() error {
	return wrapErr("DisableExclusiveMode", ioctl.Ioctl(uintptr(p.f.Load().(int)), tiocnxcl, 1))
}

// GetExclusiveMode returns true if exclusive mode is enabled for the slave pseudo-terminal device.
func (p *Port) GetExclusiveMode() (bool, error) {
	x := 0
	err := ioctl.Ioctl(uintptr(p.f.Load().(int)), tiocgexcl, uintptr(unsafe.Pointer(&x)))
	if err != nil {
		return false, wrapErr("GetExclusiveMode", err)
	}
	return x == 1, nil
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

func (attrs *Termios2) SetCustomIOSpeed(iSpeed, oSpeed uint32) {
	attrs.Cflag &= ^(CBAUD)
	attrs.Cflag |= BOTHER
	attrs.ISpeed = iSpeed
	attrs.OSpeed = oSpeed
}

func (attrs *Termios2) SetCustomSpeed(speed uint32) {
	attrs.Cflag &= ^(CBAUD)
	attrs.Cflag |= BOTHER
	attrs.ISpeed = speed
	attrs.OSpeed = speed
}
