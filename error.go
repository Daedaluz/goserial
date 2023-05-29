package serial

import "syscall"

type Error struct {
	msg string
	err error
}

func (e Error) Error() string {
	if e.msg != "" {
		msg := e.msg
		if e.err != nil {
			msg += ": " + e.err.Error()
		}
		return msg
	}
	if e.err != nil {
		return e.err.Error()
	}
	return ""
}

func (e Error) Unwrap() error {
	return e.err
}

func wrapErr(msg string, e error) error {
	if e == nil {
		return nil
	}
	return Error{
		msg: msg,
		err: e,
	}
}

var (
	ErrClosed = Error{"port already closed", syscall.EBADF}
)
