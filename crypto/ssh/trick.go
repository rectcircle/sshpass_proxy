package ssh

import (
	"errors"
	"fmt"
	"net"
)

type TrickTransport struct {
	c *connection
}

func (t *TrickTransport) WritePacket(p []byte) error {
	return t.c.transport.writePacket(p)
}
func (t *TrickTransport) ReadPacket() ([]byte, error) {
	return t.c.transport.readPacket()
}

func (t *TrickTransport) User() string {
	return t.c.User()
}

func NewServerTrickTransport(c net.Conn, config *ServerConfig) (*TrickTransport, error) {
	fullConf := *config
	fullConf.SetDefaults()
	if fullConf.MaxAuthTries == 0 {
		fullConf.MaxAuthTries = 6
	}
	// Check if the config contains any unsupported key exchanges
	for _, kex := range fullConf.KeyExchanges {
		if _, ok := serverForbiddenKexAlgos[kex]; ok {
			return nil, fmt.Errorf("ssh: unsupported key exchange %s for server", kex)
		}
	}

	s := &connection{
		sshConn: sshConn{conn: c},
	}
	_, err := s.serverHandshake(&fullConf)
	if err != nil {
		c.Close()
		return nil, err
	}
	return &TrickTransport{
		c: s,
	}, nil
}

func NewClientTrickTransport(c net.Conn, addr string, config *ClientConfig) (*TrickTransport, error) {
	fullConf := *config
	fullConf.SetDefaults()
	if fullConf.HostKeyCallback == nil {
		c.Close()
		return nil, errors.New("ssh: must specify HostKeyCallback")
	}

	conn := &connection{
		sshConn: sshConn{conn: c, user: fullConf.User},
	}

	if err := conn.clientHandshake(addr, &fullConf); err != nil {
		c.Close()
		return nil, fmt.Errorf("ssh: handshake failed: %v", err)
	}
	return &TrickTransport{
		c: conn,
	}, nil
}

func TrickTransportPacketCopy(a, b *TrickTransport) error {
	for {
		bytes, err := a.ReadPacket()
		if err != nil {
			return err
		}
		err = b.WritePacket(bytes)
		if err != nil {
			return err
		}
	}
}

func ErrIsDisconnectedByUser(err error) bool {
	if e, ok := err.(*disconnectMsg); ok && e.Reason == 11 {
		return true
	}
	return false
}
