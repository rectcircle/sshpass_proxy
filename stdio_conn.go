package sshpass_proxy

import (
	"fmt"
	"net"
	"os"
	"time"
)

type stdioConn struct{}

func wrapStdioError(inErr, outErr error) error {
	if inErr != nil && outErr != nil {
		return fmt.Errorf("stdin error: %v, stdout error: %v", inErr, outErr)
	}
	if inErr != nil {
		return inErr
	}
	if outErr != nil {
		return outErr
	}
	return nil
}

// Close implements net.Conn
func (stdioConn) Close() error {
	inErr := os.Stdin.Close()
	outErr := os.Stdout.Close()
	return wrapStdioError(inErr, outErr)
}

type stdioAddr struct{}

// Network implements net.Addr
func (stdioAddr) Network() string {
	return "stdio"
}

// String implements net.Addr
func (stdioAddr) String() string {
	return "stdio"
}

// LocalAddr implements net.Conn
func (stdioConn) LocalAddr() net.Addr {
	return stdioAddr{}
}

// Read implements net.Conn
func (stdioConn) Read(b []byte) (n int, err error) {
	return os.Stdin.Read(b)
}

// RemoteAddr implements net.Conn
func (stdioConn) RemoteAddr() net.Addr {
	return stdioAddr{}
}

// SetDeadline implements net.Conn
func (stdioConn) SetDeadline(t time.Time) error {
	inErr := os.Stdin.SetDeadline(t)
	outErr := os.Stdout.SetDeadline(t)
	return wrapStdioError(inErr, outErr)
}

// SetReadDeadline implements net.Conn
func (stdioConn) SetReadDeadline(t time.Time) error {
	return os.Stdin.SetReadDeadline(t)
}

// SetWriteDeadline implements net.Conn
func (stdioConn) SetWriteDeadline(t time.Time) error {
	return os.Stdout.SetWriteDeadline(t)
}

// Write implements net.Conn
func (stdioConn) Write(b []byte) (n int, err error) {
	return os.Stdout.Write(b)
}

func StdioConn() net.Conn {
	return stdioConn{}
}
