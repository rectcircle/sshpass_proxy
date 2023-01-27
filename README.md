# sshpass_proxy

sshpass_proxy 实现了和 [sshpass][sshpass] 类似的功能，即 SSH 连接时，自动提供密码。和 [sshpass][sshpass] 相比，sshpass_proxy 是基于 proxy 思路的实现，sshpass_proxy 命令会用在 ssh 命令的 [`ProxyCommand`](https://man.openbsd.org/ssh_config#ProxyCommand) 选项中。这样做的好处是，可以将其配置到 ssh_config 配置文件中，在 [VSCode Remote SSH](https://code.visualstudio.com/docs/remote/ssh) 场景中特别有用。

## 编译安装

```bash
CGO_ENABLED=0 go build -o sshpass_proxy ./cmd
sudo mv sshpass_proxy /usr/local/bin
```

## 使用

（下面示例，假设密码为 `123456`）

### 通过 ssh_config (推荐)

配置 `~/.ssh/config`。

```
Host demo
    User remoteUser
    HostName remoteHost
    StrictHostKeyChecking no
    ProxyCommand sshpass_proxy -p 123456 -a %h:%p  # 也可以从文件读取密码，参见下文。
```

执行 ssh 命令。

```bash
ssh demo
```

### 通过 SSH 命令

使用 -p 直接指定密码。

```bash
ssh -o "StrictHostKeyChecking=no" -o "ProxyCommand=sshpass_proxy -p 123456 -a %h:%p" remoteUser@remoteHost
```

使用环境变量 `SSHPASS` 指定密码。

```bash
export SSHPASS=123456
ssh -o "StrictHostKeyChecking=no" -o "ProxyCommand=sshpass_proxy -e -a %h:%p" remoteUser@remoteHost
```

使用文件指定密码

```bash
cat passwd-file # 输出为 123456
ssh -o "StrictHostKeyChecking=no" -o "ProxyCommand=sshpass_proxy -f passwd-file -a %h:%p" remoteUser@remoteHost
```

### 命令参数

`sshpass_proxy -h`

```
Usage of sshpass_proxy:
  -a string
        Provide target ssh server addr
  -e    Password is passed as env-var "SSHPASS"
  -f string
        Take password to use from file
  -h    Show help (this screen)
  -k string
        Provide sshpass proxy server host private key (default "/Users/sunben.96/Library/Application Support/sshpass_proxy/host_private_key.id_rsa")
  -p string
        Provide password as argument (security unwise)
```

## 注意事项

* 使用 sshpass_proxy 时，ssh 命令获取到的 remote host key 的指纹是 sshpass_proxy 的指纹，而不是真正的目标主机的主机。因此，如果先 `ssh remoteUser@remoteHost` 手动输入密码连接，再使用 sshpass_proxy 连接，将会出现 `WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!` 错误。因此上文通过 `StrictHostKeyChecking=no` 来禁用 remote host key 检查。

## 原理

参见博客： [SSH 反向代理](https://www.rectcircle.cn/posts/ssh-reverse-proxy/)。

正如上文所示，sshpass_proxy 只是是 SSH 反向代理的一个具体应用。该项目可以作为一个 SSH 反向代理库来使用，可以用来实现：SSH 跳板机等场景。

## 依赖的开源项目

* 许可证为 [BSD-3-Clause license](https://cs.opensource.google/go/x/crypto/+/master:LICENSE)  的 [golang.org/x/crypto](https://cs.opensource.google/go/x/crypto) 项目中 SSH 传输层和认证协议部分的源码。

[sshpass]: https://github.com/kevinburke/sshpass
