# aliyun-dns

借助阿里云DNS的API，通过定时检测IP变动，自动更新DNS域名解析数据来实现动态域名解析（DDNS）功能。

功能：

- [x] 支持指定配置文件
- [x] 支持IPv4和IPv6同时更新
- [x] 支持多域名
- [x] 支持宽带有多个公网IP，或者有多个宽带的情景（比如宽带多拨）
- [x] 支持缓存
- [x] 支持自定义TTL
- [x] 支持自定义IP检测周期
- [x] 支持Docker

## 简介

最近发现宽带运营商分配了IPv6地址，此前自己用Erlang写的DDNS程序并没有支持这种情况，遂计划用Go重新写一份。

本项目SDK的调用借鉴了 [airring/aliyun-dns][3]，特此感谢 @airring 。

## 一、配置

参考 [config.yaml](./config.yaml):

### 1.1、配置参数说明

- `accessKeyId`: 阿里云的授权ID
- `accessKeySecret`: 与 `accessKeyId` 对应的授权密钥
- `log`：日志相关配置
    + `path`：日志文件保存的路径，例如：`/var/log/`。
    + `level`：日志记录的最低级别，有 debug,info,warn,error 这四种。例如配置为 info 时就不会记录debug日志。
    + `develop`：开发者模式，开启后会输出代码文件和堆栈信息。
- `ipv4_check_url`: 通过该URL获取网络的IPv4地址, 当前仅支持返回IP的URL，返回json或其它复杂数据结构无法处理
- `ipv6_check_url`: 通过该URL获取网络的IPv6地址, 当前仅支持返回IP的URL，返回json或其它复杂数据结构无法处理
- `ttl`: 域名的TTL，从阿里云后台查询，普通域名是 600
- `interval`: 公网IP检测周期，单位秒(s)
- `broadband_retry`: 重复次数

    不能小于宽带多拨情景下的宽带数量（可能会出现的公网IP数量），如果配置了负载均衡，还要保证数量能覆盖到所有宽带（负载均衡的最小公倍数）。
- `customer`:
    + `domain`: 主域名（如果要更新的域名是 `dns.aliyuncs.com`，那么`aliyuncs.com` 是主域名，`dns` 是子域名前缀）
    + `ipv4_rr`: IPv4地址对应的 **子域名前缀**
    + `ipv6_rr`: IPv6地址对应的 **子域名前缀**

### 1.2、配置注意事项

- 1、 `ipv4_check_url` 和 `ipv6_check_url` 用来获取本机的公网IP，不写则不使用对应项，**至少配置一个**。
- 2、 `ipv4_rr` 和 `ipv6_rr` 不写则不使用对应项，**至少配置一个**。
- 3、 允许配置多个域名，在 `customer` 下配置多个域名时，会自动更新这些域名的解析记录。

### 1.3、阿里云授权

**accessKeyId** 和 **accessKeySecret** 需要从阿里云 用户AccessKey下拉菜单中申请。 

> 详细可以进入: [如何申请AccessKey][1]

## 二、运行

### 2.1、编译运行

```
git clone https://github.com/qingchuwudi/aliyun-dns
cd aliyun-dns
go build .
./aliyun-dns -c config.yaml
```

### 2.2、Docker运行

```bash
docker run --name=aliyun-dns --restart=unless-stopped --net=host -v /etc/alidns-config.yaml:/etc/config.yaml -d qingchuwudi/aliyun-dns:v2.1.1
```

**注意：** 
- 1、要使用 `--net=host` 网络模式启动，和宿主机使用同一个网络，这样可以保证获取的IP与宿主机一致。

  参考资料 [Docker 网络:host模式][2]
- 2、修改配置文件以后，重启docker生效：`docker restart aliyun-dns`。
- 3、支持多种不同架构环境下运行：**amd64(即x86_64)**、**386（即x86）**、**arm64**、**arm/v7**。

> Docker使用桥接或者其它网络模式运行时，Docker内的IPv6和主机真实的IPv6不一定相同。
> 
> 因为国内IPv4是路由下面的DHCP，但是观察发现IPv6是为局域网环境中每个主机分配一个公网IPv6地址，并不是一个宽带使用一个IPv6地址。

_ _ _
[1]: https://help.aliyun.com/knowledge_detail/63482.html
[2]: https://www.freeaihub.com/article/host-module-in-docker-network.html
[3]: https://github.com/airring/aliyun-dns
