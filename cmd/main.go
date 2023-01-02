package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strings"

	sshpass_proxy "github.com/rectcircle/sshpass_proxy"
	"github.com/rectcircle/sshpass_proxy/crypto/ssh"
)

type CLIParam struct {
	Password                  string
	Addr                      string
	ProxyServerHostPrivateKey ssh.Signer
}

func logParseArgsFatal(a ...interface{}) {
	os.Stderr.WriteString(fmt.Sprintln(a...))
	flag.Usage()
	os.Exit(1)
}

func generateRSA2048PrivateKeyPemToFile(file string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	privateKeyFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer privateKeyFile.Close()

	pemPrivateKey := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	err = pem.Encode(privateKeyFile, pemPrivateKey)
	if err != nil {
		return err
	}
	return nil
}

func parseArgs() *CLIParam {
	var (
		passwordFile                  string
		password                      string
		passwordFromEnv               bool
		showHelp                      bool
		proxyServerHostPrivateKeyFile string
		addr                          string
		proxyServerHostPrivateKey     ssh.Signer
	)
	defaultProxyServerHostPrivateKeyFile := "sshpass_proxy/host_private_key.id_rsa"
	userConfigDir, err := os.UserConfigDir()
	if err == nil {
		defaultProxyServerHostPrivateKeyFile = path.Join(userConfigDir, defaultProxyServerHostPrivateKeyFile)
	} else {
		defaultProxyServerHostPrivateKeyFile = path.Join(".config", defaultProxyServerHostPrivateKeyFile)
	}
	// 密码部分参数风格来自:
	// https://github.com/kevinburke/sshpass/blob/master/main.c#L79
	flag.StringVar(&passwordFile, "f", "", "Take password to use from file")
	flag.StringVar(&password, "p", "", "Provide password as argument (security unwise)")
	flag.StringVar(&addr, "a", "", "Provide target ssh server addr")
	flag.StringVar(&proxyServerHostPrivateKeyFile, "k", defaultProxyServerHostPrivateKeyFile, "Provide sshpass proxy server host private key")
	flag.BoolVar(&passwordFromEnv, "e", false, "Password is passed as env-var \"SSHPASS\"")
	flag.BoolVar(&showHelp, "h", false, "Show help (this screen)")
	flag.Parse()
	if showHelp {
		flag.Usage()
		os.Exit(0)
	}
	if addr == "" {
		logParseArgsFatal("error: param `-a addr` not found")
	}
	// 秘钥: 读取默认配置文件或自动生成
	if proxyServerHostPrivateKeyFile == "" || proxyServerHostPrivateKeyFile == defaultProxyServerHostPrivateKeyFile {
		proxyServerHostPrivateKeyFile = defaultProxyServerHostPrivateKeyFile
		_, err = os.Stat(proxyServerHostPrivateKeyFile)
		exist := true
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				exist = false
			} else {
				logParseArgsFatal("proxy server host rsa private key file stat error:", err)
			}
		}
		if !exist {
			err = os.MkdirAll(path.Dir(proxyServerHostPrivateKeyFile), 0755)
			if err != nil {
				logParseArgsFatal("init sshpass_proxy config dir error:", err)
			}
			err = generateRSA2048PrivateKeyPemToFile(proxyServerHostPrivateKeyFile)
			if err != nil {
				logParseArgsFatal("generate proxy server host rsa private key encode pem PKCS1 error:", err)
			}
		}
	}
	// 秘钥: 解析
	privateBytes, err := os.ReadFile(proxyServerHostPrivateKeyFile)
	if err != nil {
		logParseArgsFatal("read proxy server host rsa private key file error:", err)
	}
	proxyServerHostPrivateKey, err = ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		logParseArgsFatal("parse proxy server host rsa private key error:", err)
	}
	// 最高优先级是 -p password
	// 次优先级是 -f passwordFile
	if password == "" && passwordFile != "" {
		c, err := os.ReadFile(passwordFile)
		if err != nil {
			logParseArgsFatal("read password file error:", err)
		}
		password = strings.TrimSpace(string(c))
		if password == "" {
			logParseArgsFatal("error: password file is empty")
		}
	}
	// 最低优先级是 -e 读取 SSHPASS 环境变量
	if password == "" && passwordFromEnv {
		password = os.Getenv("SSHPASS")
		if password == "" {
			logParseArgsFatal("error: env SSHPASS not found")
		}
	}
	if password == "" {
		logParseArgsFatal("error: password not found")
	}
	return &CLIParam{
		Password:                  password,
		Addr:                      addr,
		ProxyServerHostPrivateKey: proxyServerHostPrivateKey,
	}
}

// ssh -o "ProxyCommand sh -c 'go run ./cmd -p 123456 -a %h:%p 2> test.log'" -p 2222 root@127.0.0.1

func main() {
	param := parseArgs()
	clientConn := sshpass_proxy.StdioConn()
	serverConn, err := net.Dial("tcp", param.Addr)
	if err != nil {
		log.Fatal("failed dial target server conn: ", err.Error())
	}
	err = sshpass_proxy.SSHPassPorxy(
		clientConn, serverConn,
		param.ProxyServerHostPrivateKey, param.Addr, param.Password)
	if err != nil {
		log.Fatal(err.Error())
	}
}
