[toc]

## upex-wallet new 


upex-wallet钱包, 用于`充值`, `提现`, `热转冷`,`归集` 操作; 批量生成`充值地址`


### 一.项目结构
```shell

```

### 二.wallet-tools(工具)

功能说明:

Tools for deposit & withdraw.

- statuschecker, running on ubuntu@127.0.0.1.
- cloneaccount
- genwallet, walletctl
- rundeck
- ecdsautil, show secp256k1 public key in compressed and uncompressed format.
- genrsa, generate rsa key pair.

#### 1.环境准备

For Ubuntu:
```shell
apt install build-essential cmake libboost-all-dev
```

For CentOS:
```shell
sudo yum groupinstall 'Development Tools'
sudo yum install centos-release-scl
sudo  yum install devtoolset-7

sudo yum install epel-release
sudo yum install cmake3

sudo yum install boost-devel
sudo yum install boost-static
sudo ln -s /usr/lib64/libboost_thread-mt.a /usr/lib64/libboost_thread.a
```

#### 2. 构建

```shell
git clone
cd ~/upex-wallet-eth-new/wallet-tools/build

make 
```

#### 3.使用
生成钱包, 充值地址, sql文件, txt文件

1. Generate wallet, address, sql file for deposit.
```shell
# 生成100个eth的普通地址和一个系统地址, 存放在当前目录下, 密码为 123456
> ./bin/genwallet -c ETH -n 100 -N 1 -d . -p 123456
> ll 
eth-deposit-address-1.sql   # 充值地址, 用于deposit程序, 关联 Address表
eth-deposit-addrs.txt       # 充值地址, 导入到交易所中数据库使用
eth-withdraw-address-1.sql  # 提现程序使用,普通地址, 关联提现数据库Address表
eth-withdraw-address-2.sql  # 提现程序使用, 系统地址, 关联提现数据库Address表
wallet.dat                  # 钱包文件, 用于签名程序
wallet.dat.meta             # 钱包文件, 用于签名    
```

2. genrsa generate rsa key pair

```shell
> ./bin/genrsa -f broadcast,signer,transfer
> ll
broadcast_rsa                       # 广播服务私钥
broadcast_rsa.public                # 广播服务公钥
signer_rsa                          # 签名私钥
signer_rsa.public                   # 签名公钥    
transfer_rsa                        # 交易公钥
transfer_rsa.public                 # 交易私钥
```

### 三. wallet-deposit (充值程序)

#### 上币配置文件修改

1. 在 `build/make/` 目录下添加相应币种的 makefile 文件。
```shll
> ls -l
total 72
-rw-r--r--  1 ubuntu  staff   100B  9 14 16:23 btc.mk
-rw-r--r--  1 ubuntu  staff    47B  9 14 16:23 eos.mk
-rw-r--r--  1 ubuntu  staff    72B  9 14 16:23 eth.mk
-rw-r--r--  1 ubuntu  staff   185B  9 14 16:23 iota.mk
-rwxr-xr-x  1 ubuntu  staff   563B  9 14 16:23 make.sh
-rw-r--r--  1 ubuntu  staff    62B  9 14 16:23 nodeapi.mk
-rw-r--r--  1 ubuntu  staff    51B  9 14 16:23 ripple.mk
-rw-r--r--  1 ubuntu  staff    90B  9 13 14:15 tool.mk
-rw-r--r--  1 ubuntu  staff    49B  9 14 16:23 usdt.mk
```
2. 修改 `build/Makefile` 文件。

```shell
build: btc eth usdt ripple iota eos
```

3. build

安装 golang 环境, 1.13版本以上,环境变量,设置好`GOMODULE=on`

```shell
cd build
make
```
或直接使用go编译
```shell
cd ../cmd/deposit
go build -o deposit-eth main.go
```

4. 运行

修改配置, 将`wallet-confi/deposit/config/app.yml`与主程序`eth`到同一目录
```shell
./deposit-eth -c app.yml
```



### 四. wallet-transfer (交易)

- syncd 区块同步
- broadcast 交易广播
- signer 交易签名
- transfer (提现, 热转冷, 归集, 资产余额检查, 告警)



#### 1. syncd 区块同步

1. 构建

   ```shell
   # 切换目录下upex-wallet-eth-new\wallet-withdraw\cmd\syncd\main.go
   go build -o syncd-eth main.go
   ```

2. 需添加如下配置—-–使用 `wallet-config/deposit/config/app.yml` 配置文件

   ```shell
   # special setting
   ignoreNotifyAudit: true           # in withdraw program,just need store deposit record, but not request broker notify
   # trxAirDropAddress: 0
   ```

3.  运行

   ```shell
   ./syncd-eth -c path/to/app.yml
   ```

#### 2. broadcast( 广播交易) 

1. 构建

   ```shell
   # 切换目录至`upex-wallet-eth-new\wallet-withdraw\cmd\broadcast`
   go build -o broadcast main.go
   ```

2. 修改配置文件— 使用(`wallet-config/withdraw/broadcast/config/app.yml`)

   ```shell
   # 添加broadcast的私钥文件, 使用wallet-tools生成
   eth:
   	rsaKey:---
   ```

3. 运行

   ```shell
   ./broadcast -c path/to/app.yml
   ```

#### 3. signer (交易签名服务)

1. 构建

   ```shell 
   # 切换目录至`upex-wallet-eth-new\wallet-withdraw\cmd\signer`
   go build -o signer main.go
   ```

2. 修改配置—使用(`wallet-config/withdraw/signer/config/app.yml`)

   ```shell
   # signer服务私钥
   rsaKey: |
     -----BEGIN RSA PRIVATE KEY-----
     -----END RSA PRIVATE KEY-----
   
   # broadcast服务公钥
   rsaPubKey: |
     -----BEGIN PUBLIC KEY-----
     -----END PUBLIC KEY-----
   ```

3.  运行

   ```shell
   # 1. 启动
   ./signer -c path/to/app.yml
   
   # 2. 输入生成 rsa密钥对时的密码前三位
   password: 123
   
   ```

#### 4. transfer (提现服务)

1. 构建

   ```shell
   # 切换目录至`upex-wallet-eth-new\wallet-withdraw\cmd\transfer`
   go build -o eth-withraw main.go
   ```

2. 修改配置 使用(`wallet-config/withdraw/transfer/config/app.yml`)

   ```shell
   sign:
     # 签名服务
     url: "http://127.0.0.1:8998"
     # 使用signer公钥对生成交易地址的密码后三位进行加密----"456"
     # 使用 `wallet-withdraw/cmd/transfer/util.go程序`
     # 加密使用
     pass: "kCUU2+lCEuN67mPK91LgCBRimV0t5BN1shHmZ9xo8HOtmRzdDvP5lITRi9WyYWEMRHWP1JZzs9DMT2KtmKQ4W8HYOGb1SRbHN9w4dkPYCcSXvIlWhzDm4bZN45VTR0FYljVE7Nxeatwe5nzx6zz0dVgyuPZAO61WiuZ4ZGAdniAOjkjdbfyPravjRTeg/nD8z6FqP2fL7AVMU910rnUzcoO7wLy0+TnzD8g9v4vd3Or4Dh285NjVxY2zipAOwfv64DnNSbUat+szCmgcT0ygmPEoVd6whfjv1RGg2mGyFojOguRwaVPJi3PZwVid2DXwUn1NapmOvoXDk8Vv4t0Jjg=="
     timeout: 5
   
   wallet:
     # broadcast 服务----注意端口
     broadcastUrl: "http://127.0.0.1:8880/v1/tx/broadcast"
     # 交易手续费配置
     maxFee: 1
     alarmTimeout: 30
     maxAccountRemain: 10  # 10 eth, grater than value, will do cool-down tx
     minAccountRemain: 6   # 6 eth , less than value, will alarm
     coldAddress: ""       # cold wallet address
   ```

3. 启动

   ```shell
   ./eth-withdraw -c path/to/app.yml
   ```

   

