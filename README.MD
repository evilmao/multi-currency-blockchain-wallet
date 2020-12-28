
## upex-wallet new 


upex-wallet钱包, 用于`充值`, `提现`, `热转冷`,`归集` 操作; 批量生成`充值地址`

### 币种支持
- `ETH`, `ERC20(ETC, SMT, SMT, IONC...)` 
- `BTC`, `BTC-LIKE`(`ETP`, `QTUM`, `FAB`, `MONA`, `LTC`)
- `TRX`, `TRC20`(`USDT`, `BTT`...)

### 一.项目结构
```shell
.
├── go.mod						# 第三方依赖包
├── go.sum						# 第三方包版本管理
├── README.MD						# README.MD
├── vendor						# Vendor本地第三方包
├── wallet-base						# wallet-base 基础组件
│   ├── api
│   ├── blockchain
│   ├── cmd
│   ├── currency
│   ├── db
│   │   └── db.go
│   ├── jsonrpc
│   ├── libs
│   ├── main.go
│   ├── models					        # 数据库表结构及方法			
│   │   ├── account.go				        # account表, 地址最终余额
│   │   ├── account_test.go			        # account测试文件
│   │   ├── address.go				        # address表,充值地址
│   │   ├── blockinfo.go				# blockinfo表,最新区块详情
│   │   ├── currency.go				        # currency表, 合约地址表
│   │   ├── models.go				        # 初始化数据库表方法
│   │   ├── tx.go					# deposit_tx表,充值记录
│   │   └── utxo.go					# utxo表,utxo交易记录
│   ├── monitor						# data-dog服务监控
│   ├── newbitx						# 依赖包
│   ├── service						# service规范
│   │   ├── service.go
│   │   └── worker.go
│   ├── util						# 工具函数
│   └── viper						# yml文件读取配置方法
├── wallet-config				        # 配置文件
│   ├── deposit						# depoist/syncd程序配置文件
│   │   └── config
│   └── withdraw					# 提现相关配置
│       ├── broadcast				        # 交易广播配置
│       ├── signer					# 交易签名配置
│       └── transfer					# 提现/热转冷/归集配置
├── wallet-deposit					# deposit主程序
│   ├── base						# 数据库表初始化
│   │   └── models
│   ├── build						# make build脚本
│   ├── cmd						# 命令行启动主程序
│   │   ├── cmd.go
│   │   ├── deposit					# deposit服务主程序main函数
│   │   └── version.go
│   ├── deposit						# 业务代码
│   │   ├── notifier.go				        # notifier服务,充值后实时与交易所进行反馈
│   │   ├── util.go
│   │   ├── util_test.go
│   │   └── worker.go			                # 区块同步,解析,入账
│   ├── rpc						# rpc相关方法
│   │   ├── blockscanrpc.go				# 定义区块扫描相关方法
│   │   ├── cache.go					# 区块缓存处理
│   │   ├── eth						# eth rpc相关方法实现								
│   │   └── rpc.go					# rpc接口定义	
│   ├── services
│   └── syncer						# 区块同步
│       ├── fetcher.go				        # 区块获取
│       └── syncer.go					# 同步逻辑
├── wallet-tools				        # wallet-tools主程序
│   ├── base
│   ├── build						# make build构建脚本
│   └── cmd
│       ├── cloneaccount
│       ├── cmdserver
│       ├── ecdsautil
│       ├── genrsa				        # 生成密钥对
│       ├── genwallet					# 生成地址,公钥,钱包文件,sql文件
│       ├── rundeck
│       └── statuschecker
└── wallet-withdraw					# withdraw主程序
    ├── base
    ├── broadcast					# 广播交易逻辑代码
    ├── build						# make build脚本
    ├── cmd						# 命令行启动程序
    │   ├── broadcast					# broadcast main入口
    │   ├── cmd.go					# 命令行解析实现
    │   ├── signer					# signer main 入口
    │   ├── syncd					# syncd main入口
    │   ├── transfer			                # transfer main入口	
    │   └── version.go
    ├── signer						#交易签名逻辑
    │   ├── client.go					# 签名机客户端调用实现逻辑
    │   ├── client_test.go
    │   ├── keystore.go				        # 签名
    │   └── server.go				        # 签名机服务端实现
    ├── syncd						# 区块同步逻辑
    │   ├── notifier.go
    │   ├── util.go
    │   └── worker.go
    └── transfer					# 提现闲逛逻辑
        ├── alarm					# 告警逻辑
        ├── checker					# 地址余额监控逻辑
        ├── cooldown				        # 热转冷转账逻辑	
        ├── gather					# 归集逻辑
        ├── helper.go			
        ├── rollback												
        ├── txbuilder		                        # 核心代码: 交易构建,打包
        ├── utxofee				        # 实时获取最佳手续费(btc)				
        └── withdraw					# 提现逻辑
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
**注意:** 所有币种都使用相同充值程序, 币种支持见[币种支持列表](#币种支持)

1. build

    安装 golang 环境, 1.13版本以上,环境变量,设置好`GOMODULE=on`
    ```shell
    cd wallet-deposit/build
    make
    ```

2. 运行
    修改配置, 将`wallet-config/deposit/config/app.yml`与主程序`deposit`到同一目录
    ```shell
    # 1. 主程序重命名命名 
    mkdir -p deposit/eth &&cp -r wallet-deposit/bin/deposit/upex_eth_deposit ./
    # 2. 修改配置文件后将配置文件复制到 主程序同一目录
    cp -r wallet-config/deposit/config/app.yml ./
    # 3. 启动
    ./upex_eth_deposit -c app.yml
    # 4. 可使用后台进程
    nohup ./upex_eth_deposit -c app.ynl >/dev/null 2>&1 &
    # 5.查看日志
    tail -100f log/wallet-deposit-eth.log
    ```

3. 注意事项
添加代币功能需要在 配置中手动加入配置后, 重启 deposit服务和 syncd服务
```shell script
  ...
    # erc20 或 trc20 配置相关代币信息
    symbols:
     usdt:                                                       # token symbol 
        precision: 6                                             # token amount decimals
        address: "0xdac17f958d2ee523a2206206994597c13d831ec7"    # token of contract address, check from Official web
        minDepositAmount: 0.2                                    # less than value, will not notify to the broker
        minBalanceRemain: 2                                      # System address balance less than this value, will send mail
        maxBalanceRemain: 4                                      # System address balance grater than this value, will do cold-down tx
```
### 四. wallet-transfer (交易)

- syncd 区块同步
- broadcast 交易广播
- signer 交易签名
- transfer (提现, 热转冷, 归集, 资产余额检查, 告警)

#### 4.1 方式一: 使用make一次性构建服务

1. 切换至 `upex-wallet-eth-new/wallet-withdraw/build`

   ```shell
   ls
   ├── make
   │   ├── build_monero.sh
   │   ├── make_gitversion.sh
   │   ├── make_import_claim.sh
   │   ├── make_import.sh
   │   └── make.sh
   └── Makefile
   ```

2. 构建

   > 非管理员权限, 需要修改`make`目录下构建脚本的权限
   >
   > ```shell
   > chmod 700 make/*
   > ```


3. 执行make指令

   ```shell
   make 
   # 如下服务
   .
   ├── bin
   │   ├── broadcast   # 广播服务
   │   ├── signer      # 签名机服务
   │   ├── syncd       # 区块同步
   │   └── transfer    # 交易服务
   ├── make
   │   ├── build_monero.sh
   │   ├── make_gitversion.sh
   │   ├── make_import_claim.sh
   │   ├── make_import.sh
   │   └── make.sh
   └── Makefile
   ```

4. 配置相应文件后, 分别启动服务


#### 4.2 方式二: 单独构建
##### 1. syncd 区块同步

1. 构建

   ```shell
   # 切换目录下upex-wallet-eth-new\wallet-withdraw\cmd\syncd\main.go
   go build -o syncd-eth main.go
   ```

2. 需添加如下配置—-–使用 `wallet-config/deposit/config/app.yml` 配置文件

   ```shell
   # 数据库使用提现数据库, 不同于充值
   dsn: "root:Xxzx@789@tcp(43.254.54.79:3306)/wallet_eth?charset=utf8&parseTime=True&loc=Local"

   # special setting
   ignoreNotifyAudit: true           # in withdraw program,just need store deposit record, but not request broker notify
   # trxAirDropAddress: 0
   ```

3. 运行

   ```shell
   ./syncd-eth -c path/to/app.yml
   ```

##### 2. broadcast( 广播交易) 

1. 构建

   ```shell
   # 切换目录至`upex-wallet-eth-new\wallet-withdraw\cmd\broadcast`
   go build -o broadcast main.go
   ```

2. 修改配置文件— 使用(`wallet-config/withdraw/broadcast/config/app.yml`)

   ```shell
   
   eth:
        maxConfirm: 30       # 链上交易网络节点确认数
        verifyInterval: 10   # 链上交易 网络请求间隔, 单位秒
   	rsaKey:---			 # broadcast的私钥文件, 使用wallet-tools生成
   ```
3. 运行

   ```shell
   ./broadcast -c path/to/app.yml
   ```
##### 3. signer (交易签名服务)

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

3. 运行
   ```shell
   # 1. 启动
   ./signer -c path/to/app.yml
   
   # 2. 输入生成 rsa密钥对时的密码前三位
   password: 123
   ```

##### 4. transfer (提现服务)

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
     # 使用 `wallet-withdraw/cmd/transfer/util/util.go 脚本`
     # 加密使用
     pass: "kCUU2+lCEuN67mPK91LgCBRimV0t5BN1shHmZ9xo8HOtmRzdDvP5lITRi9WyYWEMRHWP1JZzs9DMT2KtmKQ4W8HYOGb1SRbHN9w4dkPYCcSXvIlWhzDm4bZN45VTR0FYljVE7Nxeatwe5nzx6zz0dVgyuPZAO61WiuZ4ZGAdniAOjkjdbfyPravjRTeg/nD8z6FqP2fL7AVMU910rnUzcoO7wLy0+TnzD8g9v4vd3Or4Dh285NjVxY2zipAOwfv64DnNSbUat+szCmgcT0ygmPEoVd6whfjv1RGg2mGyFojOguRwaVPJi3PZwVid2DXwUn1NapmOvoXDk8Vv4t0Jjg=="
     timeout: 5
   
   wallet:
     # broadcast 服务----注意端口
     broadcastUrl: "http://127.0.0.1:8880/v1/tx/broadcast"
     minFee: 0.0003                 # 最低交易手续费配置 0.003 eth
     maxFee: 0.0144                 # 最高交易手续费配置 0.0144 eth
     maxAccountRemain: 1  		 	 # 系统地址最大预留金额 1 eth
     minAccountRemain: 0.8   		 # 系统地址小于预留金额会触发报警 0.8eth
     coldAddress: ""       		 # 冷钱包地址
     coolDownTaskInterval: 15       # 人转冷时间间隔, 单位分钟
     #alarmTimeout: 30
   ```

3. 启动
   ```shell
   ./eth-withdraw -c path/to/app.yml
   ```

### 其他

#### 邮件告警
用来监控 `提现`, `热转冷` ,`归集` 交易故障时告警; `账户余额不足`,`热转冷资金变动`通知.

1. 修改配置文件 `wallet-config/withdraw/transfer/app.yml`
```shell script
email:
  errAlarmInterval: 15                      # same error trigger sending email interval, units `minutes`
  host: "smtp.gmail.com"                    # email host
  port: "465"                               # ssl port
  from: "fc.notify.1@gmail.com"             # from email address
  password: "wgcfkddggccddtwj"              # dedicated password(not account password)
  to: "faily0117@gmail.com"                 # to email address
```

2. 使用HTML模板文件发送邮件: 将模板文件移动至主程序同已目录下
```shell script
├── withdraw						
│   ├── transfer                    # transfer 主程序
│   ├── app.yml                     # transfer config    
│   ├── email-template.html         # email HTML 模板
...
```

   

