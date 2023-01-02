// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"errors"
	"fmt"
	"net"
	"time"
)

// clientHandshake performs the client side key exchange. See RFC 4253 Section
// 7.
func (c *connection) clientHandshake(dialAddress string, config *ClientConfig) error {
	if config.ClientVersion != "" {
		c.clientVersion = []byte(config.ClientVersion)
	} else {
		c.clientVersion = []byte(packageVersion)
	}
	var err error
	c.serverVersion, err = exchangeVersions(c.sshConn.conn, c.clientVersion)
	if err != nil {
		return err
	}

	c.transport = newClientTransport(
		newTransport(c.sshConn.conn, config.Rand, true /* is client */),
		c.clientVersion, c.serverVersion, config, dialAddress, c.sshConn.RemoteAddr())
	if err := c.transport.waitSession(); err != nil {
		return err
	}

	c.sessionID = c.transport.getSessionID()
	return c.clientAuthenticate(config)
}

// verifyHostKeySignature verifies the host key obtained in the key exchange.
// algo is the negotiated algorithm, and may be a certificate type.
func verifyHostKeySignature(hostKey PublicKey, algo string, result *kexResult) error {
	sig, rest, ok := parseSignatureBody(result.Signature)
	if len(rest) > 0 || !ok {
		return errors.New("ssh: signature parse error")
	}

	if a := underlyingAlgo(algo); sig.Format != a {
		return fmt.Errorf("ssh: invalid signature algorithm %q, expected %q", sig.Format, a)
	}

	return hostKey.Verify(result.H, sig)
}

// HostKeyCallback is the function type used for verifying server
// keys.  A HostKeyCallback must return nil if the host key is OK, or
// an error to reject it. It receives the hostname as passed to Dial
// or NewClientConn. The remote address is the RemoteAddr of the
// net.Conn underlying the SSH connection.
type HostKeyCallback func(hostname string, remote net.Addr, key PublicKey) error

// BannerCallback is the function type used for treat the banner sent by
// the server. A BannerCallback receives the message sent by the remote server.
type BannerCallback func(message string) error

// A ClientConfig structure is used to configure a Client. It must not be
// modified after having been passed to an SSH function.
type ClientConfig struct {
	// Config contains configuration that is shared between clients and
	// servers.
	Config

	// User contains the username to authenticate as.
	User string

	// Auth contains possible authentication methods to use with the
	// server. Only the first instance of a particular RFC 4252 method will
	// be used during authentication.
	Auth []AuthMethod

	// HostKeyCallback is called during the cryptographic
	// handshake to validate the server's host key. The client
	// configuration must supply this callback for the connection
	// to succeed. The functions InsecureIgnoreHostKey or
	// FixedHostKey can be used for simplistic host key checks.
	HostKeyCallback HostKeyCallback

	// BannerCallback is called during the SSH dance to display a custom
	// server's message. The client configuration can supply this callback to
	// handle it as wished. The function BannerDisplayStderr can be used for
	// simplistic display on Stderr.
	BannerCallback BannerCallback

	// ClientVersion contains the version identification string that will
	// be used for the connection. If empty, a reasonable default is used.
	ClientVersion string

	// HostKeyAlgorithms lists the public key algorithms that the client will
	// accept from the server for host key authentication, in order of
	// preference. If empty, a reasonable default is used. Any
	// string returned from a PublicKey.Type method may be used, or
	// any of the CertAlgo and KeyAlgo constants.
	HostKeyAlgorithms []string

	// Timeout is the maximum amount of time for the TCP connection to establish.
	//
	// A Timeout of zero means no timeout.
	Timeout time.Duration
}
