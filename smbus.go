// Copyright 2017 The go-daq Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package smbus provides access to the System Management bus, over i2c.
//
// http://www.smbus.org/.
package smbus

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	i2cSlave      = 0x0703
	i2cSlaveForce = 0x0706
	i2cFuncs      = 0x0705
	i2cSMBus      = 0x0720

	i2cSMBusWrite uint8 = 0
	i2cSMBusRead  uint8 = 1

	// size identifiers
	i2cSMBusByteData     uint32 = 2
	i2cSMBusWordData     uint32 = 3
	i2cSMBusBlockData    uint32 = 5
	i2cSMBusI2CBlockData uint32 = 8
	i2cSMBusBlockMax     uint32 = 32
)

// Conn is connection to a i2c device.
type Conn struct {
	f *os.File
}

// Open opens a connection to the i2c bus number at address addr.
func Open(bus int, addr uint8) (*Conn, error) {
	f, err := os.OpenFile(fmt.Sprintf("/dev/i2c-%d", bus), os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := ioctl(f.Fd(), i2cSlave, uintptr(addr)); err != nil {
		return nil, err
	}
	return &Conn{f: f}, nil
}

// Write sends buf to the remote i2c device.
// The interpretation of the message is implementation dependant.
func (c *Conn) Write(buf []byte) (int, error) {
	return c.f.Write(buf)
}

// WriteByte sends a single byte to the remote i2c device.
// The interpretation of the message is implementation dependant.
func (c *Conn) WriteByte(b byte) (int, error) {
	var buf [1]byte
	buf[0] = b
	return c.f.Write(buf[:])
}

// Read reads data from the remote i2c device into p.
func (c *Conn) Read(p []byte) (int, error) {
	return c.f.Read(p)
}

// Close closes the connection to the remote i2c device.
func (c *Conn) Close() error {
	return c.f.Close()
}

// ReadReg reads a single byte from a designated register.
func (c *Conn) ReadReg(addr, reg uint8) (uint8, error) {
	if err := c.addr(addr); err != nil {
		return 0, err
	}

	var v uint8
	cmd := i2cCmd{
		rw:  i2cSMBusRead,
		cmd: reg,
		len: i2cSMBusByteData,
		ptr: unsafe.Pointer(&v),
	}
	ptr := unsafe.Pointer(&cmd)
	err := ioctl(c.f.Fd(), i2cSMBus, uintptr(ptr))
	return v, err
}

// WriteReg writes a single byte v to a designated register.
func (c *Conn) WriteReg(addr, reg, v uint8) error {
	if err := c.addr(addr); err != nil {
		return err
	}

	cmd := i2cCmd{
		rw:  i2cSMBusWrite,
		cmd: reg,
		len: i2cSMBusByteData,
		ptr: unsafe.Pointer(&v),
	}
	ptr := unsafe.Pointer(&cmd)
	return ioctl(c.f.Fd(), i2cSMBus, uintptr(ptr))
}

// ReadWord reads a 2-bytes word from a designated register.
func (c *Conn) ReadWord(addr, reg uint8) (uint16, error) {
	if err := c.addr(addr); err != nil {
		return 0, err
	}

	var v uint16
	cmd := i2cCmd{
		rw:  i2cSMBusRead,
		cmd: reg,
		len: i2cSMBusWordData,
		ptr: unsafe.Pointer(&v),
	}
	ptr := unsafe.Pointer(&cmd)
	err := ioctl(c.f.Fd(), i2cSMBus, uintptr(ptr))
	return v, err
}

// WriteWord writes a 2-bytes word v to a designated register.
func (c *Conn) WriteWord(addr, reg uint8, v uint16) error {
	if err := c.addr(addr); err != nil {
		return err
	}

	cmd := i2cCmd{
		rw:  i2cSMBusWrite,
		cmd: reg,
		len: i2cSMBusWordData,
		ptr: unsafe.Pointer(&v),
	}
	ptr := unsafe.Pointer(&cmd)
	return ioctl(c.f.Fd(), i2cSMBus, uintptr(ptr))
}

func (c *Conn) addr(addr uint8) error {
	return ioctl(c.f.Fd(), i2cSlave, uintptr(addr))
}

func ioctl(fd, cmd, arg uintptr) (err error) {
	_, _, e1 := syscall.Syscall6(syscall.SYS_IOCTL, fd, cmd, arg, 0, 0, 0)
	if e1 != 0 {
		err = e1
	}
	return
}

type i2cCmd struct {
	rw  uint8
	cmd uint8
	len uint32
	ptr unsafe.Pointer
}