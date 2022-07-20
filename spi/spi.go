package spi

import (
	ioctl "github.com/daedaluz/goioctl"
	"reflect"
	"syscall"
	"unsafe"
)

const spi_ioc_magic = 'k'

type spi_ioc_transfer struct {
	txBuf uint64
	rxBuf uint64

	len      uint32
	speed_hz uint32

	delay_usecs      uint16
	bits_per_word    uint8
	cs_change        uint8
	tx_nbits         uint8
	rx_nbits         uint8
	word_delay_usecs uint8
	pad              uint8
}

var (
	// Read / Write of SPI mode (SPI_MODE_0..SPI_MODE_3) (limited to 8 bits)
	spi_ioc_rd_mode = ioctl.IOR(spi_ioc_magic, 1, 1)
	spi_ioc_wr_mode = ioctl.IOW(spi_ioc_magic, 1, 1)

	// Read / Write SPI bit justification
	spi_ioc_rd_lsb_first = ioctl.IOR(spi_ioc_magic, 2, 1)
	spi_ioc_wr_lsb_first = ioctl.IOW(spi_ioc_magic, 2, 1)

	// Read / Write SPI device word length (1..N)
	spi_ioc_rd_bits_per_word = ioctl.IOR(spi_ioc_magic, 3, 1)
	spi_ioc_wr_bits_per_word = ioctl.IOW(spi_ioc_magic, 3, 1)

	// Read / Write SPI device default max speed hz
	spi_ioc_rd_max_speed_hz = ioctl.IOR(spi_ioc_magic, 4, 4)
	spi_ioc_wr_max_speed_hz = ioctl.IOW(spi_ioc_magic, 4, 4)

	// Read / Write of the SPI mode field
	spi_ioc_rd_mode32 = ioctl.IOR(spi_ioc_magic, 5, 4)
	spi_ioc_wr_mode32 = ioctl.IOW(spi_ioc_magic, 5, 4)

	spi_ioc_message = ioctl.IOW(spi_ioc_magic, 0, unsafe.Sizeof(spi_ioc_transfer{}))
)

type Mode uint32

type Device struct {
	fd  int
	cfg *Config
}

type Config struct {
	Mode          Mode
	Bits          uint8
	Speed         uint32
	DelayUsec     uint16
	CSChange      bool
	TXNBits       uint8
	RXNBits       uint8
	WordDelayUsec uint8
}

func (d *Device) Write(data []byte) (n int, err error) {
	return syscall.Write(d.fd, data)
}

func (d *Device) Tx(data []byte) (read []byte, err error) {
	read = make([]byte, len(data))

	dataHeader := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	readHeader := (*reflect.SliceHeader)(unsafe.Pointer(&read))

	xferBlock := &spi_ioc_transfer{
		txBuf:            uint64(dataHeader.Data),
		rxBuf:            uint64(readHeader.Data),
		len:              uint32(dataHeader.Len),
		speed_hz:         d.cfg.Speed,
		delay_usecs:      d.cfg.DelayUsec,
		bits_per_word:    d.cfg.Bits,
		cs_change:        0,
		tx_nbits:         d.cfg.TXNBits,
		rx_nbits:         d.cfg.RXNBits,
		word_delay_usecs: d.cfg.WordDelayUsec,
		pad:              0,
	}
	if d.cfg.CSChange {
		xferBlock.cs_change = 1
	}
	err = ioctl.Ioctl(d.fd, spi_ioc_message, uintptr(unsafe.Pointer(xferBlock)))
	return
}

func (d *Device) Close() error {
	return syscall.Close(d.fd)
}

func Open(path string, cfg *Config) (*Device, error) {
	fd, err := syscall.Open(path, syscall.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	if err := ioctl.Ioctl(fd, spi_ioc_wr_max_speed_hz, uintptr(unsafe.Pointer(&cfg.Speed))); err != nil {
		syscall.Close(fd)
		return nil, err
	}

	if err := ioctl.Ioctl(fd, spi_ioc_wr_bits_per_word, uintptr(unsafe.Pointer(&cfg.Bits))); err != nil {
		syscall.Close(fd)
		return nil, err
	}

	if err := ioctl.Ioctl(fd, spi_ioc_wr_mode32, uintptr(unsafe.Pointer(&cfg.Mode))); err != nil {
		syscall.Close(fd)
		return nil, err
	}
	dev := &Device{
		fd:  fd,
		cfg: cfg,
	}
	return dev, nil
}
