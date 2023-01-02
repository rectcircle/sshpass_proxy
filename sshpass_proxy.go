package sshpass_proxy

import (
	"fmt"
	"net"

	"github.com/rectcircle/sshpass_proxy/crypto/ssh"
)

// SSHPassPorxy 功能类似于 sshpass，但是原理完全不同。
//
// SSHPassPorxy 通过一个 ssh 传输层协议 proxy，实现 openssh 的客户端，可以对使用密码鉴权的 ssh server 实现免密登录。
//
//                                                           SSHPassProxy
//  +--------+           +------------------------------------------------------------------------------+          +--------+
//  | client | ---> (clientConn) ssh transport server <--- Packet Copy ---> ssh transport client (serverConn) ---> | server |
//  +--------+           +------------------------------------------------------------------------------+          +--------+
func SSHPassPorxy(
	clientConn, serverConn net.Conn,
	proxyServerHostPrivateKey ssh.Signer,
	serverAddr, serverPassword string,
) error {
	// 1. 使用 ssh server 对接来 client 连接，完成握手和免密鉴权，并获取到 ssh 传输层协议对象。
	proxyServerConfig := &ssh.ServerConfig{
		NoClientAuth: true,
	}
	proxyServerConfig.AddHostKey(proxyServerHostPrivateKey)
	proxyServerTransport, err := ssh.NewServerTrickTransport(clientConn, proxyServerConfig)
	if err != nil {
		return fmt.Errorf("failed handshake and authenticate with sshpass proxy server: %v", err)
	}
	serverUser := proxyServerTransport.User()
	// 2. 使用 ssh client 对接 server 连接，完成握手和密码鉴权，并获取到 ssh 传输层协议对象。
	proxyClientConfig := &ssh.ClientConfig{
		User: serverUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(serverPassword),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
	}
	proxyClientTransport, err := ssh.NewClientTrickTransport(serverConn, serverAddr, proxyClientConfig)
	if err != nil {
		return fmt.Errorf("failed handshake and authenticate with target server: %v", err)
	}
	// 转发
	errc := make(chan error, 1)
	go func() {
		errc <- ssh.TrickTransportPacketCopy(proxyServerTransport, proxyClientTransport)
	}()
	go func() {
		errc <- ssh.TrickTransportPacketCopy(proxyClientTransport, proxyServerTransport)
	}()
	if err = <-errc; err != nil && !ssh.ErrIsDisconnectedByUser(err) {
		return err
	}
	return nil
}
