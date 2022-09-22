## 基本概念

### 一、什么是防火墙？

在计算机中，防火墙是基于预定安全规则来监视和控制传入和传出网络流量的网络安全系统。该计算机流入流出的所有网络通信均要经过此防火墙。防火墙对流经它的网络通信进行扫描，这样能够过滤掉一些攻击，以免其在目标计算机上被执行。防火墙还可以关闭不使用的端口。而且它还能禁止特定端口的流出通信，封锁特洛伊木马。最后，它可以禁止来自特殊站点的访问，从而防止来自不明入侵者的所有通信。



#### 1.1 防火墙分为软件防火墙和硬件防火墙，他们的优缺点：

**硬件防火墙**：拥有经过特别设计的硬件及芯片，性能高、成本高(当然硬件防火墙也是有软件的，只不过有部分功能由硬件实现，所以硬件防火墙其实是硬件+软件的方式)；

**软件防火墙**：应用软件处理逻辑运行于通用硬件平台之上的防火墙，性能比硬件防火墙低、成本低。



#### 1.2 Netfilter与iptables的关系

Netfilter是由Rusty Russell提出的Linux 2.4内核防火墙框架，该框架既简洁又灵活，可实现安全策略应用中的许多功能，如数据包过滤、数据包处理、地址伪装、透明代理、动态网络地址转换(Network Address Translation，NAT)，以及基于用户及媒体访问控制(Media Access Control，MAC)地址的过滤和基于状态的过滤、包速率限制等。Iptables/Netfilter的这些规则可以通过灵活组合，形成非常多的功能、涵盖各个方面，这一切都得益于它的优秀设计思想。

Netfilter是Linux操作系统核心层内部的一个数据包处理模块，它具有如下功能：

- 网络地址转换(Network Address Translate)

- 数据包内容修改

- 以及数据包过滤的防火墙功能

Netfilter平台中制定了数据包的五个挂载点(Hook Point，我们可以理解为回调函数点，数据包到达这些位置的时候会主动调用我们的函数，使我们有机会能在数据包路由的时候改变它们的方向、内容)，这5个挂载点分别是`PRE_ROUTING`、`INPUT`、`OUTPUT`、`FORWARD`、`POST_ROUTING`。

Netfilter所设置的规则是存放在内核空间中的，而**iptables是一个应用层的应用程序，它通过Netfilter放出的接口来对存放在内核空间中的 XXtables(Netfilter的配置表)进行修改**。这个XXtables由表tables、链chains、规则rules组成，iptables在应用层负责修改这个规则文件，类似的应用程序还有firewalld(CentOS7默认防火墙)。

所以Linux中真正的防火墙是Netfilter，但由于都是通过应用层程序如iptables或firewalld进行操作，所以我们一般把iptables或firewalld叫做Linux的防火墙。

**注意**：以上说的iptables都是针对IPv4的，如果IPv6，则要用ip6tables，至于用法应该是跟iptables是一样的。

注：Linux系统运行时，内存分内核空间和用户空间，内核空间是Linux内核代码运行的空间，它能直接调用系统资源，用户空间是指运行用户程序的空间，用户空间的程序不能直接调用系统资源，必须使用内核提供的接口"system call"。



### 二、链的概念

iptables开启后，数据报文从进入服务器到出来会经过5道关卡，分别为Prerouting(路由前)、Input(输入)、Outpu(输出)、Forward(转发)、Postrouting(路由后)：

![数据流转](https://doc.xujianqq.com.cn/doc/03/ae2ec6f73b5dd9849c93dfc074cac997.jpg)

每一道关卡中有多个规则，数据报文必须按顺序一个一个匹配这些规则，这些规则串起来就像一条链，所以我们把这些关卡都叫`链`：

![chain](https://doc.xujianqq.com.cn/doc/03/e3ffc9afdb8cd8bd450ed917e3fe50e3.png)

- **INPUT链**：当接收到防火墙本机地址的数据包(入站)时，应用此链中的规则；

- **OUTPUT链**：当防火墙本机向外发送数据包(出站)时，应用此链中的规则；

- **FORWARD链**：当接收到需要通过防火墙发送给其他地址的数据包(转发)时，应用此链中的规则；

- **PREROUTING链**：在对数据包作路由选择之前，应用此链中的规则，如DNAT；

- **POSTROUTING链**：在对数据包作路由选择之后，应用此链中的规则，如SNAT。

其中中INPUT、OUTPUT链更多的应用在"主机防火墙"中，即主要针对服务器本机进出数据的安全控制；而FORWARD、PREROUTING、POSTROUTING链更多的应用在"网络防火墙"中，特别是防火墙服务器作为网关使用时的情况。

### 三、表的概念

虽然每一条链上有多条规则，但有些规则的作用(功能)很相似，多条具有相同功能的规则合在一起就组成了一个`表`，iptables提供了四种`表`：

- **filter表**：主要用于对数据包进行过滤，根据具体的规则决定是否放行该数据包(如DROP、ACCEPT、REJECT、LOG)，所谓的防火墙其实基本上是指这张表上的过滤规则，对应内核模块iptables_filter；

- **nat表**：network address translation，网络地址转换功能，主要用于修改数据包的IP地址、端口号等信息(网络地址转换，如SNAT、DNAT、MASQUERADE、REDIRECT)。属于一个流的包(因为包的大小限制导致数据可能会被分成多个数据包)只会经过这个表一次，如果第一个包被允许做NAT或Masqueraded，那么余下的包都会自动地被做相同的操作，也就是说，余下的包不会再通过这个表。对应内核模块iptables_nat；

- **mangle表**：拆解报文，做出修改，并重新封装，主要用于修改数据包的TOS(Type Of Service，服务类型)、TTL(Time To Live，生存周期)指以及为数据包设置Mark标记，以实现Qos(Quality Of Service，服务质量)调整以及策略路由等应用，由于需要相应的路由设备支持，因此应用并不广泛。对应内核模块iptables_mangle；

- **raw表**：是自1.2.9以后版本的iptables新增的表，主要用于决定数据包是否被状态跟踪机制处理，在匹配数据包时，raw表的规则要优先于其他表，对应内核模块iptables_raw。

**我们最终定义的防火墙规则，都会添加到这四张表中的其中一张表中。**

### 四、表链关系

5条链(即5个关卡)中，并不是每条链都能应用所有类型的表，事实上除了Ouptput链能同时有四种表，其他链都只有两种或三种表：

![table](https://doc.xujianqq.com.cn/doc/03/55afd069e4f1ba01d87cee0b9322c6c7.png)

实际上由上图我们可以看出，无论在哪条链上，raw表永远在mangle表上边，而mangle表永远在nat表上边，nat表又永远在filter表上边，这表明各表之间是有匹配顺序的。

前面说过，数据报文必须按顺序匹配每条链上的一个一个的规则，但其实同一类(即属于同一种表)的规则是放在一起的，不同类的规则不会交叉着放，按上边的规律，每条链上各个表被匹配的顺序为：`raw → mangle → nat → filter`。

我们最终定义的防火墙规则，都会添加到这四张表中的其中一张表中，所以我们实际操作是对`表`进行操作的，所以我们反过来说一下，每种表都能用于哪些链：

![image-20220403175755180](https://doc.xujianqq.com.cn/doc/03/image-20220403175755180.png)

综上，数据包通过防火墙的流程可总结为下图：

![protocol](https://doc.xujianqq.com.cn/doc/03/ac586d71025972c3c200ca6bc96917c5.png)


### 五、规则的概念

iptables规则主要包含`条件&动作`，即匹配出符合什么条件(规则)后，对它采取怎样的动作。

#### 5.1 匹配条件

```shell
-i --in-interface    网络接口名     指定数据包从哪个网络接口进入，
-o --out-interface   网络接口名     指定数据包从哪个网络接口输出
-p ---proto          协议类型       指定数据包匹配的协议，如TCP、UDP和ICMP等
-s --source          源地址或子网   指定数据包匹配的源地址
   --sport           源端口号       指定数据包匹配的源端口号
   --dport           目的端口号     指定数据包匹配的目的端口号
-m --match           匹配的模块      指定数据包规则所使用的过滤模块
```

#### 5.2 处理的动作

iptables处理动作除了 `ACCEPT、REJECT、DROP、REDIRECT 、MASQUERADE `以外，还多出 `LOG、ULOG、DNAT、RETURN、TOS、SNAT、MIRROR、QUEUE、TTL、MARK`等。我们只说明其中最常用的动作：

- ACCEPT 允许数据包通过

- REJECT 拦阻该数据包，并返回数据包通知对方，可以返回的数据包有几个选择：ICMP port-unreachable、ICMP echo-reply 或是tcp-reset（这个数据包包会要求对方关闭联机），进行完此处理动作后，将不再比对其它规则，直接中断过滤程序。范例如下：

```shell
$ iptables -A  INPUT -p TCP --dport 22 -j REJECT --reject-with ICMP echo-reply
```

- DROP 丢弃数据包不予处理，进行完此处理动作后，将不再比对其它规则，直接中断过滤程序。
- REDIRECT 将封包重新导向到另一个端口（PNAT），进行完此处理动作后，将会继续比对其它规则。这个功能可以用来实作透明代理 或用来保护web 服务器。例如：

```shell
$ iptables -t nat -A PREROUTING -p tcp --dport 80 -j REDIRECT--to-ports 8081
```

- MASQUERADE 改写封包来源IP为防火墙的IP，可以指定port 对应的范围，进行完此处理动作后，直接跳往下一个规则链（mangle:postrouting）。这个功能与 SNAT 略有不同，当进行IP 伪装时，不需指定要伪装成哪个 IP，IP 会从网卡直接读取，当使用拨接连线时，IP 通常是由 ISP 公司的 DHCP服务器指派的，这个时候 MASQUERADE 特别有用。范例如下：

```shell
$ iptables -t nat -A POSTROUTING -p TCP -j MASQUERADE --to-ports 21000-31000
```

- LOG 将数据包相关信息纪录在 /var/log 中，详细位置请查阅 /etc/syslog.conf 配置文件，进行完此处理动作后，将会继续比对其它规则。例如：

```shell
$ iptables -A INPUT -p tcp -j LOG --log-prefix `input packet`
```

- SNAT 改写封包来源 IP 为某特定 IP 或 IP 范围，可以指定 port 对应的范围，进行完此处理动作后，将直接跳往下一个规则炼（mangle:postrouting）。范例如下：

```
$ iptables -t nat -A POSTROUTING -p tcp-o eth0 -j SNAT --to-source 192.168.10.15-192.168.10.160:2100-3200
```

- DNAT 改写数据包包目的地 IP 为某特定 IP 或 IP 范围，可以指定 port 对应的范围，进行完此处理动作后，将会直接跳往下一个规则链（filter:input 或 filter:forward）。范例如下：

```shell
$ iptables -t nat -A PREROUTING -p tcp -d 15.45.23.67 --dport 80 -j DNAT --to-destination 192.168.10.1-192.168.10.10:80-100
```

- MIRROR 镜像数据包，也就是将来源 IP与目的地IP对调后，将数据包返回，进行完此处理动作后，将会中断过滤程序。
- QUEUE 中断过滤程序，将封包放入队列，交给其它程序处理。透过自行开发的处理程序，可以进行其它应用，例如：计算联机费用等。
- RETURN 结束在目前规则链中的过滤程序，返回主规则链继续过滤，如果把自订规则炼看成是一个子程序，那么这个动作，就相当于提早结束子程序并返回到主程序中。
- MARK 将封包标上某个代号，以便提供作为后续过滤的条件判断依据，进行完此处理动作后，将会继续比对其它规则。范例如下：

```shell
$ iptables -t mangle -A PREROUTING -p tcp --dport 22 -j MARK --set-mark 22
```

其中REJECT和DROP有点类似，以下是服务器设置REJECT和DROP后，ping这个服务器的响应的区别：

**REJECT动作：**

```shell
PING 10.37.129.9 (10.37.129.9): 56 data bytes
92 bytes from centos-linux-6.5.host-only (10.37.129.9): Destination Port Unreachable
Vr HL TOS  Len   ID Flg  off TTL Pro  cks      Src      Dst
 4  5  00 5400 29a3   0 0000  40  01 3ab1 10.37.129.2  10.37.129.9
 
Request timeout for icmp_seq 0
92 bytes from centos-linux-6.5.host-only (10.37.129.9): Destination Port Unreachable
Vr HL TOS  Len   ID Flg  off TTL Pro  cks      Src      Dst
 4  5  00 5400 999d   0 0000  40  01 cab6 10.37.129.2  10.37.129.9
```

**DROP动作**：

```shell
PING 10.37.129.9 (10.37.129.9): 56 data bytes
Request timeout for icmp_seq 0
Request timeout for icmp_seq 1
Request timeout for icmp_seq 2
Request timeout for icmp_seq 3
Request timeout for icmp_seq 4
```

### 六、iptables命令操作

对iptables进行操作，其实就是对它的四种`表`进行`增删改查`操作。

#### 6.1 启动iptables

由于国内大部分公司的服务器都采用CentOS系统，所以这里以CentOS为例。CentOS6及其之前的CentOS系统都默认使用iptables防火墙，但CentOS7默认已经没有iptables防火墙了，取而代之的是firewalld，不过你还是可以自己安装iptables。

##### 6.1.1 对于CentOS6：

启动iptables(事实上应该叫`加载`iptables模块更合适，前面也说过iptables并不是一个真正的服务，它没有进程，这意味着你用`ps -ef | grep iptables`是看不到进程的)：

```shell
$ service iptables start
```

其它命令：

```shell
# 查看启动状态
$ service iptables status
# 停止iptables
$ service iptables stop
# 重启iptables(重启其实就是先stop再start)
$ service iptables restart
# 重载就是重新加载配置的规则，在这里貌似跟重启一样
$ service iptables reload
```

##### 6.1.2 对于CentOS7：

前面说过，CentOS7默认没有iptables(CentOS7默认防火墙是firewalld)，但我们还是可以自己安装iptables。

CentOS7关闭firewalld防火墙：

```shell
$ systemctl stop firewalld
```

CentOS7关闭firewalld防火墙开机启动：

```shell
$ systemctl disable firewalld
```

CentOS7安装iptables防火墙：

```shell
$ sudo yum -y install iptables
```

CentOS7安装iptables的service启动工具：

```shell
$ sudo yum -y install iptables-services
```

安装完之后，所有启动、停止之类的操作与CentOS6都一样，参考前面CentOS6的操作即可。

但实际上在CentOS7中用service来操作程序的启停，它会自动跳转到使用systemctl，所以其实除了用CentOS6的方法，还可以用CentOS7的：

```shell
# 启动iptables
$ systemctl start iptables
# 查看iptables状态
$ systemctl status iptables
# 停止iptables
$ systemctl stop iptables
# 重启iptables
$ systemctl restart iptables
# 重载iptables
$ systemctl reload iptables
```

这里要特别注意，稍微了解Linux的童鞋都知道，service命令是用于启动Linux的进程的，但在这里是例外，`service iptables start`并没有启动一个进程，你无法用`ps aux | grep iptables`的方式看到一个叫iptables的进程，你只能用`service iptables status`去查看它的状态。

所以iptables其实不能叫`服务`，因为它并没有一个`守护进程`，其实iptables只是相当于一个客户端工具，真正的防火墙是Linux内核中的netfilter，由于netfilter是内核功能，用户无法直接操作，iptables这个工具是提供给用户设置过滤规则的，但最终这个过滤规则是由netfilter来执行的。

### 七、查询规则

命令格式：`iptables [选项] [参数]`

#### 7.1 常用选项：

`-L`: list的缩写，list我们通常翻译成列表，意思是列出每条链上的规则，因为多条规则就是一个列表，所以用`-L`来表示。`-L`后面还可以跟上5条链`(POSTROUTING、INPUT、FORWARD、OUTPUT、POSTROUTING)`的其中一条链名，注意链名必须全大写，如查看`INPUT链`上的规则:

```shell
$ iptables -L INPUT
Chain INPUT (policy ACCEPT)
target     prot opt source               destination
ACCEPT     all  --  anywhere             anywhere            state RELATED,ESTABLISHED
ACCEPT     icmp --  anywhere             anywhere
ACCEPT     all  --  anywhere             anywhere
ACCEPT     tcp  --  anywhere             anywhere            state NEW tcp dpt:ssh
REJECT     all  --  anywhere             anywhere            reject-with icmp-host-prohibited
```

不指定的话就是默认查看所有链上的规则列表:

```shell
$ iptables -L
Chain INPUT (policy ACCEPT)
target     prot opt source               destination
ACCEPT     all  --  anywhere             anywhere            state RELATED,ESTABLISHED
ACCEPT     icmp --  anywhere             anywhere
ACCEPT     all  --  anywhere             anywhere
ACCEPT     tcp  --  anywhere             anywhere            state NEW tcp dpt:ssh
REJECT     all  --  anywhere             anywhere            reject-with icmp-host-prohibited
Chain FORWARD (policy ACCEPT)
target     prot opt source               destination
REJECT     all  --  anywhere             anywhere            reject-with icmp-host-prohibited
Chain OUTPUT (policy ACCEPT)
target     prot opt source               destination
```

- **Chain INPUT**: INPUT链上的规则，同理，后面的`Chain FORWARD`、`Chain OUTPUT`分别是FORWARD链和OUTPUT链上的规则；

- **(policy ACCEPT)**: 表示默认策略是接受，即假如我没设置，那就是允许，只有我设置哪个不允许，才会不允许，示例中是安装iptables后的默认规则，由于默认是ACCEPT，你规则也设置为ACCEPT按道理来说是没什么意义的，因为你不设置也是ACCEPT呀，但事实上，是为了方便修改为REJECT/DROP等规则，说白了就是放在那，要设置的时候我们就可以直接修改；

- **target**: 英文意思是`目标`，但该列的值通常是动作，比如ACCEPT(接受)、REJECT(拒绝)等等，但它确实可以是`目标`，比如我们创建 一条链`iptables -N July_filter`，然后在INPUT链上添加一条规则，让它跳转到刚刚的新链`-A INPUT -p tcp -j July_filter`，再用`iptables -L`查看，可以看到target此时已经是真正的`target(July_filter)`而不再是动作了：

```shell
Chain INPUT (policy ACCEPT)
target     prot opt source               destination
ACCEPT     all  --  anywhere             anywhere            state RELATED,ESTABLISHED
ACCEPT     icmp --  anywhere             anywhere
ACCEPT     all  --  anywhere             anywhere
ACCEPT     tcp  --  anywhere             anywhere            state NEW tcp dpt:ssh
REJECT     all  --  anywhere             anywhere            reject-with icmp-host-prohibited
July_filter  tcp  --  anywhere             anywhere
Chain FORWARD (policy ACCEPT)
target     prot opt source               destination
REJECT     all  --  anywhere             anywhere            reject-with icmp-host-prohibited
Chain OUTPUT (policy ACCEPT)
target     prot opt source               destination
Chain July_filter (1 references)
target     prot opt source               destination
```

- **prot**: protocol，协议；

- **opt**: option，选项；

- **source**: 源地址(ip(可以是网段)/域名/主机名)

- **destination**: 目标地址(ip(可以是网段)/域名/主机名)

- **末列**: 一些额外的信息

`-t`：前面`-L`不是列出所有链的规则列表吗？为什么没有PREROUTING和POSTROUTING链呢？因为有默认参数`-t`，t是table的缩写，意思是指定显示哪张`表`中的规则(前面说过iptables有四种表)，`iptables -L`其实就相当于`iptables -t filter -L`，即相当于你查看的是`filter`表中的规则。而根据前面的讲解，filter表只可用于INPUT、FORWARD、OUTPUT三条链中，这就是为什么`iptables -L`不显示PREROUTING链和POSTROUTING链的原因。

`-n`: numeric的缩写，numeric意思是数字的，数值的，意思是指定源和目标地址、端口什么的都以数字/数值的方式显示，否则默认会以域名/主机名/程序名等显示，该选项一般与`-L`合用，因为单独的`-n`是没有用的(没有-L列表都不显示，所以用-n就没有意义了)。
```shell
Chain INPUT (policy ACCEPT)
target     prot opt source               destination
ACCEPT     all  --  0.0.0.0/0            0.0.0.0/0           state RELATED,ESTABLISHED
ACCEPT     icmp --  0.0.0.0/0            0.0.0.0/0
ACCEPT     all  --  0.0.0.0/0            0.0.0.0/0
ACCEPT     tcp  --  0.0.0.0/0            0.0.0.0/0           state NEW tcp dpt:22
REJECT     all  --  0.0.0.0/0            0.0.0.0/0           reject-with icmp-host-prohibited

Chain FORWARD (policy ACCEPT)
target     prot opt source               destination
REJECT     all  --  0.0.0.0/0            0.0.0.0/0           reject-with icmp-host-prohibited
Chain OUTPUT (policy ACCEPT)
target     prot opt source               destination
```

其中`dpt:22`中的`dpt`是指`destination port(目标端口)`，同理，`spt`就是`source port(源端口)`。



`-v`: 基本上有点Linux常识的童鞋就应该知道，-v在Linux命令里，一般都是指`verbose`，这个词的意思是是`冗余的，啰嗦的`，即输出更加详细的信息，在iptables这里也是这个意思，一般可以跟`-L`连用：
```shell
$ iptables -v -L
Chain INPUT (policy ACCEPT 0 packets, 0 bytes)
 pkts bytes target     prot opt in     out     source               destination
13627 1033K ACCEPT     all  --  any    any     anywhere             anywhere            state RELATED,ESTABLISHED
    0     0 ACCEPT     icmp --  any    any     anywhere             anywhere
    0     0 ACCEPT     all  --  lo     any     anywhere             anywhere
    2   128 ACCEPT     tcp  --  any    any     anywhere             anywhere            state NEW tcp dpt:ssh
  275 53284 REJECT     all  --  any    any     anywhere             anywhere            reject-with icmp-host-prohibited

Chain FORWARD (policy ACCEPT 0 packets, 0 bytes)
 pkts bytes target     prot opt in     out     source               destination
    0     0 REJECT     all  --  any    any     anywhere             anywhere            reject-with icmp-host-prohibited

Chain OUTPUT (policy ACCEPT 12506 packets, 1485K bytes)
 pkts bytes target     prot opt in     out     source               destination
```

可以看到多了四列：
- pkts: packets，包的数量；
- bytes: 流过的数据包的字节数；
- in: 入站网卡；
- out: 出站网卡。

当然还可以跟前面的`-n`合用, 这样source和destination中用域名或者字符串表示的方式就换成ip了：
```shell
$ iptables -n -v -L
Chain INPUT (policy ACCEPT 0 packets, 0 bytes)
 pkts bytes target     prot opt in     out     source               destination
13642 1034K ACCEPT     all  --  *      *       0.0.0.0/0            0.0.0.0/0           state RELATED,ESTABLISHED
    0     0 ACCEPT     icmp --  *      *       0.0.0.0/0            0.0.0.0/0
    0     0 ACCEPT     all  --  lo     *       0.0.0.0/0            0.0.0.0/0
    2   128 ACCEPT     tcp  --  *      *       0.0.0.0/0            0.0.0.0/0           state NEW tcp dpt:22
  275 53284 REJECT     all  --  *      *       0.0.0.0/0            0.0.0.0/0           reject-with icmp-host-prohibited

Chain FORWARD (policy ACCEPT 0 packets, 0 bytes)
 pkts bytes target     prot opt in     out     source               destination
    0     0 REJECT     all  --  *      *       0.0.0.0/0            0.0.0.0/0           reject-with icmp-host-prohibited

Chain OUTPUT (policy ACCEPT 12516 packets, 1489K bytes)
 pkts bytes target     prot opt in     out     source               destination
```

稍微了解Linux命令的童鞋应该都知道，在很多情况下，Linux的选项是可以合并的，比如前面的`iptables -n -v -L`其实是可以合并成`iptables -nvL`的，并且参数顺序一般情况下是无关紧要的，比如`iptables -vnL`也是一样的。但是`-L`一定要写在最后，原因是`-L`是要接收参数的选项(虽然可以不传参数)，而`-v`和`-n`是不需要接收参数的，假如你写成`iptables -Lvn`，那就表示是用`-n`来接收参数了，这肯定是不行的。

`-x`: 加了`-v`后，Policy那里变成了`(policy ACCEPT 0 packets, 0 bytes)`，即多了过滤的数据包数量和字节数，其中的字节数，如果数据大了之后，会自动转换单位，比如够KB不够MB，它会显示`xxxk`，够了MB它显示`MB`，但单位转换之后，就不完全精确了，因为它没有小数，如果还是想要看以`字母`即`byptes`为单位查看的话，加个-x就行了，`x`来自于`exact`，意思是`精确的；准确的`，不取首字母应该是太多选项首字母是e了。

`--line-numbers`: 如果你想列表有序号，可以加上该选项：
```shell
$ iptables -nvL --line-numbers
```
结果中多了一列`num`，就是序号：
```shell
Chain INPUT (policy ACCEPT 0 packets, 0 bytes)
num   pkts bytes target     prot opt in     out     source               destination
1    14739 1123K ACCEPT     all  --  *      *       0.0.0.0/0            0.0.0.0/0           state RELATED,ESTABLISHED
2        0     0 ACCEPT     icmp --  *      *       0.0.0.0/0            0.0.0.0/0
3        0     0 ACCEPT     all  --  lo     *       0.0.0.0/0            0.0.0.0/0
4        2   128 ACCEPT     tcp  --  *      *       0.0.0.0/0            0.0.0.0/0           state NEW tcp dpt:22
5      305 55948 REJECT     all  --  *      *       0.0.0.0/0            0.0.0.0/0           reject-with icmp-host-prohibited
6        0     0 July_filter  tcp  --  *      *       0.0.0.0/0            0.0.0.0/0

Chain FORWARD (policy ACCEPT 0 packets, 0 bytes)
num   pkts bytes target     prot opt in     out     source               destination
1        0     0 REJECT     all  --  *      *       0.0.0.0/0            0.0.0.0/0           reject-with icmp-host-prohibited

Chain OUTPUT (policy ACCEPT 310 packets, 42772 bytes)
num   pkts bytes target     prot opt in     out     source               destination

Chain July_filter (1 references)
num   pkts bytes target     prot opt in     out     source               destination
```

其实，`--line-numbers`并不用写全，不然也太长了，其实写成`--line`就行了，甚至你不写最后一个e，即写成`--lin`都行。

合并使用各选项的命令示例：
```shell
$ iptables --line -t filter -nvxL INPUT
```

### 八、添加规则
我们可以向某条链中的某个表的最前面添加记录(我们叫`插入`，会用到-I选项，I表示Input)，也可以向某条链中的某个表的最后面添加记录(我们叫`追加`，会用到`-A`选项，A表示Append)，熟悉`vi/vim`的童鞋会对这个`I`和`A`感觉到熟悉，因为在`vi/vim`的命令模式下，按I是在光标所行的行首插入，按A是在光标所在行的行尾插入，跟这个在表头跟表尾插入非常像。

之所以有向前添加和向后添加，是因为如果前面规则的是丢弃或拒绝，那么后面的规则是不会起作用的；而如果前面的是接受后面的是丢弃或拒绝，则接受之后后面的丢弃或拒绝也是不会生效的。
向INPUT链的filter表中添加一条规则：
```shell
$ iptables -t filter -I INPUT -s 10.37.129.2 -j DROP
```
- -t: 是指定插入到哪个表中，不写的话默认为`filter`表；
- -I: 指定插入到哪条链中，并且会在该链指定表(在这里是filter表)中的最前面插入(I:Input)，如果用-A则是在最后插入(A:Append)。
- -s: 匹配源ip，s: source，源。
- -j: jump，跳转的意思，后面可指定跳转的target(目标)，比如自定义的链，当然更多的是跳转到`action(动作)`中，比如ACCEPT、DROP、REJECT等等。

整个意思，就是向iptables中的`INPUT`链`(-I INPUT)`的`filter`表`(-t filter)`的最前面`(-I)`添加一条记录，这次记录会匹配源地址为`10.39.129.2`的请求`(-s 10.39.129.2)`，并把该请求丢弃掉`(-j DROP)`。


### 8.1 删除iptables中的记录
#### 8.1.1 根据编号删除：
前面说过，查询iptables规则列表时，添加`--line-numbers`简写成`--line`即可显示记录编号，我们现在就可以根据这个编号来删除了：
```shell
$ iptables -t filter -D INPUT 2
```
`-t filte`r指定操作的表为filter表，`-D`表示delete，后面跟的两个参数，第一个是链名，第二个是要删除的规则的编号。

#### 8.1.2 根据条件删除：
```shell
$ iptables -t filter -D INPUT -s 10.37.129.2 -j DROP
```
删除INPUT链中的filter表中源地址为`10.37.129.2`并且动作为`DROP`的规则。

#### 8.1.3 清空：
`-F`: flush的缩写，flush是`冲洗、冲掉`的意思，在这里是清空的意思，`iptables -t filter -F INPUT`代表清空`INPUT`链中`filter`表中的所有规则，如果不指定链不指定表，即直接用`iptables -F`，则清空所有链中所有表的规则。

### 九、修改规则
事实上用`替换`来描述会更好一点，因为所谓的修改其实就是把整个规则替换成新的规则：
```shell
$ iptables -t filter -R INPUT 1 -s 10.37.129.3 -j ACCEPT
```
其中的`-R`就是replace，即替换的意思，整句命令意思是从INPUT链中的filter表中替换编号为1的规则，编号1后面的`-s 10.37.129.3 -j ACCEPT`就是要替换成的新规则。

**修改策略(policy):**
```shell
$ iptables -P FORWARD DROP
```

`-P`: policy，即策略。整个意思是把FORWARD链的默认规则设置为DROP，需要注意的是，策略对整个链起作用，不会在同一条链中对两个不同的表起作用，虽然`man iptables`中有`iptables [-t table] -P chain target`这个说明，但我认为这是错的，不信你可以执行以下两条命令试试，不管先执行哪行，后面执行的总会替换前面执行的(即使它们指定的表不一样)，这就说明指定表跟不指定表是没区别的，因为默认规则是作用于整条链的，无法单独对表作用：
```shell
$ iptables -t raw -P OUTPUT ACCEPT
$ iptables -t filter -P OUTPUT DROP
```

注：这个结论还有待证明，我目前认为是这样。

### 十、保存规则
#### 10.1 对于CentOS6:
```shell
$ service iptables save
```

保存时它会输出：
```shell
$ iptables: Saving firewall rules to /etc/sysconfig/iptables:[  OK  ]
```

是的，它是保存到`/etc/sysconfig/iptables`文件中的。

另一种方式保存方式：`iptables-save`命令能把需要保存规则数据输出到控制台，由前面可知，保存`iptables`规则其实是保存到`/etc/sysconfig/iptables`文件中，所以我们只要把`iptables-save`输出的内容重定向输入到`/etc/sysconfig/iptables`文件中即可：
```shell
$ iptables-save > /etc/sysconfig/iptables
```

同理，还有另一个命令用于重载配置(即以下操作相当于`service iptables reload`)：
```shell
$ iptables-restore < /etc/sysconfig/iptables
```



#### 10.2 对于CentOS7:
前面说过CentOS7需要自己安装iptables-services：
```shell
$ yum install -y iptables-services
```
安装后，就可以跟CentOS6一样保存了：
```shell
$ service iptables save
```
但要注意的是，这个save必须用service的方式，不能用systemctl，也就是不能这样：
```shell
$ systemctl save iptables
```

在CentOS7中`start/status/stop/restart/reload`都可以换成systemctl的方式，但唯独save不能换成systemctl的方式。
另外，iptables-save和iptables-restore也跟CentOS6一样可以使用。

#### 10.3 更详细的命令
`-d`：destination，用于匹配报文的目标地址，可以同时指定多个ip(逗号隔开，逗号两侧都不允许有空格)，也可指定ip段：
```shell
$ iptables -t filter -I OUTPUT -d 192.168.1.111,192.168.1.118 -j DROP
$ iptables -t filter -I INPUT -d 192.168.1.0/24 -j ACCEPT
$ iptables -t filter -I INPUT ! -d 192.168.1.0/24 -j ACCEPT
```

`-p`：用于匹配报文的协议类型,可以匹配的协议类型`tcp、udp、udplite、icmp、esp、ah、sctp`等（centos7中还支持icmpv6、mh）：
```shell
$ iptables -t filter -I INPUT -p tcp -s 192.168.1.146 -j ACCEPT
# 感叹号表示`非`，即除了匹配这个条件的都ACCEPT，但匹配这个条件不一定就是REJECT或DROP？这要看是否有为它特别写一条规则，如果没有写就会用默认策略：
$ iptables -t filter -I INPUT ! -p udp -s 192.168.1.146 -j ACCEPT
```

`-i`：用于匹配报文是从哪个网卡接口流入本机的，由于匹配条件只是用于匹配报文流入的网卡，所以在OUTPUT链与POSTROUTING链中不能使用此选项：
```shell
$ iptables -t filter -I INPUT -p icmp -i eth0 -j DROP
$ iptables -t filter -I INPUT -p icmp ! -i eth0 -j DROP
```



`-o`：用于匹配报文将要从哪个网卡接口流出本机，于匹配条件只是用于匹配报文流出的网卡，所以在INPUT链与PREROUTING链中不能使用此选项。
```shell
$ iptables -t filter -I OUTPUT -p icmp -o eth0 -j DROP
$ iptables -t filter -I OUTPUT -p icmp ! -o eth0 -j DROP
```

### 十一、iptables扩展模块
#### 11.1 tcp扩展模块
`-p tcp -m tcp --sport`用于匹配tcp协议报文的源端口，可以使用冒号指定一个连续的端口范围(`-protocol`，`-m:match`,指匹配的模块，很多人可能以为是module的缩写，其实是match的缩写，`--sport: source port`)；
`-p tcp -m tcp --dport`用于匹配tcp协议报文的目标端口，可以使用冒号指定一个连续的端口范围(`--dportestination port`)：
```shell
#示例如下
$ iptables -t filter -I OUTPUT -d 192.168.1.146 -p tcp -m tcp --sport 22 -j REJECT
$ iptables -t filter -I INPUT -s 192.168.1.146 -p tcp -m tcp --dport 22:25 -j REJECT
$ iptables -t filter -I INPUT -s 192.168.1.146 -p tcp -m tcp --dport :22 -j REJECT
$ iptables -t filter -I INPUT -s 192.168.1.146 -p tcp -m tcp --dport 80: -j REJECT
$ iptables -t filter -I OUTPUT -d 192.168.1.146 -p tcp -m tcp ! --sport 22 -j ACCEPT
```

此外，tcp扩展模块还有`–tcp-flags`选项，它可以根据TCP头部的`标识位`来匹配，具体直接点进去看吧。

#### 11.2 multiport扩展模块
`-p tcp -m multiport --sports`用于匹配报文的源端口，可以指定离散的多个端口号,端口之间用`逗号`隔开;
`-p udp -m multiport --dports`用于匹配报文的目标端口，可以指定离散的多个端口号，端口之间用`逗号`隔开：

```shell
#示例如下
$ iptables -t filter -I OUTPUT -d 192.168.1.146 -p udp -m multiport --sports 137,138 -j REJECT
$ iptables -t filter -I INPUT -s 192.168.1.146 -p tcp -m multiport --dports 22,80 -j REJECT
$ iptables -t filter -I INPUT -s 192.168.1.146 -p tcp -m multiport ! --dports 22,80 -j REJECT
$ iptables -t filter -I INPUT -s 192.168.1.146 -p tcp -m multiport --dports 80:88 -j REJECT
$ iptables -t filter -I INPUT -s 192.168.1.146 -p tcp -m multiport --dports 22,80:88 -j REJECT
```

#### 11.3 iprange扩展模块

使用iprange扩展模块可以指定`一段连续的IP地址范围`，用于匹配报文的源地址或者目标地址。iprange扩展模块中有两个扩展匹配条件可以使用：
- `--src-range`(匹配源地址范围)
- `--dst-range`(匹配目标地址范围)
```shell
$ iptables -t filter -I INPUT -m iprange --src-range 192.168.1.127-192.168.1.146 -j DROP
```

#### 11.4 string扩展模块
假设我们访问的是`http://192.168.1.146/index.html`，当`index.html`中包括`XXOO`字符时，就会被以下规则匹配上：
```shell
$ iptables -t filter -I INPUT -m string --algo bm --string `XXOO` -j REJECT
```
- `-m string`：表示使用string模块
- `--algo bm`：表示使用bm算法来匹配index.html中的字符串，`algo`是`algorithm`的缩写，另外还有一种算法叫`kmp`，所以`--algo`可以指定两种值，bm或kmp，貌似是bm算法速度比较快。

#### 11.5 time扩展模块
我们可以通过time扩展模块，根据时间段区匹配报文，如果报文到达的时间在指定的时间范围以内，则符合匹配条件。
我想要自我约束，每天早上9点到下午6点不能看网页：
```shell
$ iptables -t filter -I INPUT -p tcp --dport 80 -m time --timestart 09:00:00 --timestop 18:00:00 -j REJECT
$ iptables -t filter -I INPUT -p tcp --dport 443 -m time --timestart 09:00:00 --timestop 18:00:00 -j REJECT
```

周六日不能看网页：
```shell
$ iptables -t filter -I INPUT -p tcp --dport 80 -m time --weekdays 6,7 -j REJECT
$ iptables -t filter -I INPUT -p tcp --dport 443 -m time --weekdays 6,7 -j REJECT
```

`--weekdays`可用1-7表示一周的7天，还能用星期的缩写来指定匹配：Mon、Tue、Wed、Thu、Fri、Sat、Sun。

匹配每月22，23号：
```shell
$ iptables -t filter -I INPUT -p tcp --dport 80 -m time --monthdays 22,23 -j REJECT
```
多个条件是`相与`的关系，比如以下规则指定匹配每周五，并且这个周五还必须是在22,23,24,25,26,27,28中的一天(其实就相当于设置每月的第四个星期五)：
```shell
$ iptables -t filter -I INPUT -p tcp --dport 80 -m time --weekdays 5 --monthdays 22,23,24,25,26,27,28 -j REJECT
```

另外还有daystart和daystop：
```shell
$ iptables -t filter -I INPUT -p tcp --dport 80 -m time --daystart 2019-07-20 --daystop 2019-07-25 -j REJECT
```
`--monthdays`和`--weekdays`可以用感叹号!取反，其他的不行。

#### 11.6 connlimit模块
使用connlimit扩展模块，可以限制每个IP地址同时链接到server端的链接数量，注意：我们不用指定IP，其默认就是针对`每个客户端IP`，即对单IP的并发连接数限制。
限制22端口(ssh默认端口)连接数量上限不能超过2个；
```shell
$ iptables -t filter -I INPUT -p tcp --dport 22 -m connlimit --connlimit-above 2 -j REJECT
```

在CentOS6中可对--connlimit-above取反：
```shell
$ iptables -t filter -I INPUT -p tcp --dport 22 -m connlimit ! --connlimit-above 2 -j REJECT
```
表示连接数量只要不超过两个就允许连接，至于超过两个并不一定不允许连接，这得看默认策略是ACCEPT还是DROP或REJECT，又或者有其它规则对它进行限制。
在CentOS7中有一个叫`--connlimit-upto`的选项，它的作用跟! `--connlimit-above`一样，不过这种用法还是比较少用的。

配合`--connlimit-mask`来限制网段：
```shell
$ iptables -t filter -I INPUT -p tcp --dport 22 -m connlimit --connlimit-above 2 --connlimit-mask 24 -j REJECT
```

网址由32位二进制组成，最大可写成：255.255.255.255，而mask就是掩码(网络知识，请自行了解)，24表示24个1，即255.255.255.0。

#### 11.7 limit扩展模块

limit模块是限速用的，用于限制`单位时间内流入的数据包的数量`。

每6位秒放行一下ping包(因为1分钟是60秒，所以1分钟10个包，就相当于每6秒1个包)：
```shell
$ iptables -t filter -I INPUT -p icmp -m limit --limit 10/minite -j ACCEPT
```
`--limit`后面的单位除了minite，还可以是second、hour、day

`--limit-burst`： burst是爆发、迸发的意思，在这里是指最多允许一次性有几个包通过，要理解burst，先看以下的`令牌桶算法`。

#### 11.8 令牌桶算法：

有一个木桶，木桶里面放了5块令牌，而且这个木桶最多也只能放下5块令牌，所有报文如果想要出关入关，都必须要持有木桶中的令牌才行，这个木桶有一个神奇的功能，就是每隔6秒钟会生成一块新的令牌，如果此时，木桶中的令牌不足5块，那么新生成的令牌就存放在木桶中，如果木桶中已经存在5块令牌，新生成的令牌就无处安放了，只能溢出木桶（令牌被丢弃），如果此时有5个报文想要入关，那么这5个报文就去木桶里找令牌，正好一人一个，于是他们5个手持令牌，快乐的入关了，此时木桶空了，再有报文想要入关，已经没有对应的令牌可以使用了，但是，过了6秒钟，新的令牌生成了，此刻，正好来了一个报文想要入关，于是，这个报文拿起这个令牌，就入关了，在这个报文之后，如果很长一段时间内没有新的报文想要入关，木桶中的令牌又会慢慢的积攒了起来，直到达到5个令牌，并且一直保持着5个令牌，直到有人需要使用这些令牌，这就是令牌桶算法的大致逻辑。

看完了`令牌桶算法`，其实--limit就相当于指定`多长时间生成一个新令牌`，而`--limit-burst`则用于指定木桶中最多存放多少块令牌。

#### 11.9 udp扩展模块

udp扩展模块中能用的匹配条件比较少，只有两个，就是`--sport`与`--dport`，即匹配报文的源端口与目标端口。

放行samba服务的137和138端口：
```shell
$ iptables -t filter -I INPUT -p udp -m udp --dport 137 -j ACCEPT
$ iptables -t filter -I INPUT -p udp -m udp --dport 138 -j ACCEPT
```

当使用扩展匹配条件时，如果未指定扩展模块，iptables会默认调用与-p对应的协议名称相同的模块，所以，当使用`-p udp`时，可以省略`-m udp`：
```shell
$ iptables -t filter -I INPUT -p udp --dport 137 -j ACCEPT
$ iptables -t filter -I INPUT -p udp --dport 138 -j ACCEPT
```

udp扩展中的--sport与--dport与tcp一样，同样支持指定一个连续的端口范围：
```shell
$ iptables -t filter -I INPUT -p udp --dport 137:157 -j ACCEPT
```

`--dport 137:157`表示匹配137-157之间的所有端口。

另外与tcp一样，udp也能使用--multiport指定多个不连续的端口。

#### 11.10 icmp扩展模块
ping是使用icmp协议的，假设要禁止所有icmp协议的报文进入本机(根据前面所说，我们可以省略用`-m icmp`来指定使用icmp模块，因为不指定它会默认使用`-p`指定的协议对应的模块)：
```shell
$ iptables -t filter -I INPUT -p icmp -j REJECT
```

上述命令能产生两个效果：
- 1、别人ping本机时，无法ping通，因为数据报文无法进入；
- 2、本机ping别人时，虽然数据包可以出去，但别人的响应包也是icmp协议，无法进来(即`有去无回`)。

所以这样设置会导致，不止别人ping不通本机，本机也ping不通别人。

很明显上边的规则不是我们想要的，我们想要的一般都是允许本机ping别人，不允许别人ping本机：
```shell
$ iptables -t filter -I INPUT -p icmp --icmp-type 8/0 -j REJECT
```

`--icmp-type 8/0`用于匹配报文type为8，code为0时才会被匹配到，至于会是type和code，这是icmp协议的知识，可以参考这里iptables详解（7）：iptables扩展之udp扩展与icmp扩展。

其实上边的命令还可以省略code(即把`8/0`写成`8`即可，省略掉`/0`，原因是type=8的报文中只有code=0一种，所以我们不写默认就是code=0，不会有其它值)：
```shell
$ iptables -t filter -I INPUT -p icmp --icmp-type 8 -j REJECT
```

除了能用type/code来匹配icmp报文，还可以使用icmp的描述名称来匹配：
```shell
$ iptables -t filter -I INPUT -p icmp --icmp-type `echo-request` -j REJECT
```

`--icmp-type echo-request`的效果与`icmp --icmp-type 8/0`或`icmp --icmp-type 8`的效果完全一样(你可能发现了，icmp协议的描述`echo-request`其实是`echo request`，只不过我们用于作为匹配条件时，要把空格换成横杠)。

#### 11.11 state扩展模块

在TCP/IP协议簇中，UDP和ICMP是没有所谓的连接的，但是对于state模块来说，tcp报文、udp报文、icmp报文都是有连接状态的，我们可以这样认为，对于state模块而言，只要两台机器在`你来我往`的通信，就算建立起了连接。

而`连接`中的报文可以分为5种状态，报文状态可以为NEW、ESTABLISHED、RELATED、INVALID、UNTRACKED。具体请查看：iptables详解（8）：iptables扩展模块之state扩展

放行RELATED和ESTABLISHED状态的报文：
```shell
$ iptables -t filter -I INPUT -m state --state RELATED, ESTABLISHED -j ACCEPT
```

### 十二、其它

#### 12.1 黑白名单机制

报文在经过iptables的链时，会匹配链中的规则，遇到匹配的规则时，就执行对应的动作，如果链中的规则都无法匹配到当前报文，则使用链的默认策略（默认动作），链的默认策略通常设置为ACCEPT或者DROP。

那么，当链的默认策略设置为ACCEPT时，如果对应的链中没有配置任何规则，就表示接受所有的报文，如果对应的链中存在规则，但是这些规则没有匹配到报文，报文还是会被接受。

同理，当链的默认策略设置为DROP时，如果对应的链中没有配置任何规则，就表示拒绝所有报文，如果对应的链中存在规则，但是这些规则没有匹配到报文，报文还是会被拒绝。

所以，当链的默认策略设置为ACCEPT时，按照道理来说，我们在链中配置规则时，对应的动作应该设置为DROP或者REJECT，为什么呢？

因为默认策略已经为ACCEPT了，如果我们在设置规则时，对应动作仍然为ACCEPT，那么所有报文都会被放行了，因为不管报文是否被规则匹配到都会被ACCEPT，所以就失去了访问控制的意义。

所以，当链的默认策略为ACCEPT时，链中的规则对应的动作应该为DROP或者REJECT，表示只有匹配到规则的报文才会被拒绝，没有被规则匹配到的报文都会被默认接受，这就是`黑名单`机制。

同理，当链的默认策略为DROP时，链中的规则对应的动作应该为ACCEPT，表示只有匹配到规则的报文才会被放行，没有被规则匹配到的报文都会被默认拒绝，这就是`白名单`机制。

如果使用白名单机制，我们就要把所有人都当做坏人，只放行好人。

如果使用黑名单机制，我们就要把所有人都当成好人，只拒绝坏人。

白名单机制更加安全一些，黑名单机制更加灵活一些。

#### 12.2 自定义链

在最前面的时候就说过，iptables有五个`关卡`，即五条`链`，这些都是默认的，但其实我们可以创建自己的自定义链。

还记得前面介绍iptables -L时的那个`target`列吗？为什么target列都是一些`动作`呢？这样的话为什么不把target写成`action`呢？其实就是因为target不一定是`动作`，它还可以是`自定义链`，当指定target为自定义链时，如果匹配上了，那么就会跳转到指定的自定义链中。

你可能会问，前面一直用默认链不也都能实现想要实现的功能吗？为什么还要自定义呢？

原因是当默认链中的规则非常多时，不方便我们管理。

想象一下，如果INPUT链中存放了200条规则，这200条规则有针对httpd服务的，有针对sshd服务的，有针对私网IP的，有针对公网IP的，假如，我们突然想要修改针对httpd服务的相关规则，难道我们还要从头看一遍这200条规则，找出哪些规则是针对httpd的吗？这显然不合理。

所以，iptables中，可以自定义链，通过自定义链即可解决上述问题。

假设，我们自定义一条链，链名叫IN_WEB，我们可以将所有针对80端口的入站规则都写入到这条自定义链中，当以后想要修改针对web服务的入站规则时，就直接修改IN_WEB链中的规则就好了，即使默认链中有再多的规则，我们也不会害怕了，因为我们知道，所有针对80端口的入站规则都存放在IN_WEB链中，同理，我们可以将针对sshd的出站规则放入到OUT_SSH自定义链中，将针对Nginx的入站规则放入到IN_NGINX自定义链中，这样，我们就能想改哪里改哪里，再也不同担心找不到规则在哪里了。

但是要注意的是，自定义链并不能直接使用，而是需要被默认链引用才能够使用，即默认生效的还是默认的五条链，而自定义链必须在某条默认链的某个规则里设置target为自定义链，然后才会被引用。

创建一条名叫`IN_WEB`的自定义链：
```shell
$ iptables -N IN_WEB
Chain IN_WEB (0 references)
target     prot opt source               destination
```

我们可以看到有`0 references`这个字样，reference是`引用`的意思，`0 references`表示引用计数为0，`引用`就是前面说的`自定义链必须在某条默认链的某个规则里设置target为自定义链，然后才会被引用`。

创建一条引用`IN_WEB`链的规则(所谓的引用就是用-j跳转到该规则里)：
```shell
$ iptables -I INPUT -p tcp --dport 80 -j IN_WEB
```
此时我们再用`iptables -L`查看，可以看到`引用计数`已经是1了：
```shell
Chain IN_WEB (1 references)
target     prot opt source               destination
```


修改自定义链名称(把名为IN_WEB的自定义链的名称改为WEB，-E是edit的意思)：
```shell
$ iptables -E IN_WEB WEB
```

能修改肯定也能删除，但删除是有条件的：
- 1、自定义链没有被任何默认链引用，即自定义链的引用计数为0；
- 2、自定义链中没有任何规则，即自定义链为空。

删除名为IN_WEB的自定义链：
```shell
$ iptables -X IN_WEB
```