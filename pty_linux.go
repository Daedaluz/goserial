package serial

// OpenPTY finds an available pseudoterminal and returns a master and slave port.
// If termp is non-nil, the slave port will be configured with the given termios.
// If winp is non-nil, the slave port will be configured with the given window size.
func OpenPTY(termp *Termios, winp *Winsize) (*Port, *Port, error) {
	master, err := Open("/dev/ptmx", nil)
	if err != nil {
		return nil, nil, err
	}
	if err := master.SetLockPT(false); err != nil {
		master.Close()
		return nil, nil, err
	}
	slave, err := master.GetPTPeer(0)
	if err != nil {
		master.Close()
		return nil, nil, err
	}
	if termp != nil {
		if err := slave.SetAttr(TCSANOW, termp); err != nil {
			master.Close()
			return nil, nil, err
		}
	}
	if winp != nil {
		if err := slave.SetWinSize(winp); err != nil {
			master.Close()
			return nil, nil, err
		}
	}

	return master, slave, nil
}
