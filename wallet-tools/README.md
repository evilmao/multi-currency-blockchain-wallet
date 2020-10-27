# wallet-tools(Äã)


Tools for deposit & withdraw.

- statuschecker, running on ubuntu@127.0.0.1.
- cloneaccount
- genwallet, walletctl
- rundeck
- ecdsautil, show secp256k1 public key in compressed and uncompressed format.
- genrsa, generate rsa key pair.

## Dependency of walletctl & genwallet

For Ubuntu:
```
apt install build-essential cmake libboost-all-dev
```

For CentOS:
```
sudo yum groupinstall 'Development Tools'
sudo yum install centos-release-scl
sudo  yum install devtoolset-7

sudo yum install epel-release
sudo yum install cmake3

sudo yum install boost-devel
sudo yum install boost-static
sudo ln -s /usr/lib64/libboost_thread-mt.a /usr/lib64/libboost_thread.a
```

Build xmr keypair:
```
git clone https://github.com/fb996de/deprecated-pywallet.git ~/go/src/github.com/fb996de/wallet-tools
cd ~/go/src/github.com/fb996de/wallet-tools/build

make 
or
buildxmr=1 make
```

## Generate wallet, address, sql file

1. Generate wallet, address, sql file for deposit.

```
# generate btc wallet, 3 normal address and 2 system address, with 'fcoin' tag.
> ./bin/genwallet -c BTC -n 3 -N 2 -d ./btc/ -t fcoin
> ll ./btc/
btc-deposit-address-1-fcoin.sql
btc-deposit-addrs-1-fcoin.txt
btc-withdraw-address-1-fcoin.sql
btc-withdraw-address-2-fcoin.sql
wallet.dat
wallet.dat.meta
```

2. Generate sql file for the exchange.

```
> ./bin/cloneaccount -s "./btc/btc-deposit-addrs-1-fcoin.txt" -q "./btc/btc-ex-account-1-fcoin.sql" -i 100 -k "the-pubkey"
> ll ./btc/
btc-deposit-address-1-fcoin.sql
btc-deposit-addrs-1-fcoin.txt
btc-ex-account-1-fcoin.sql
btc-withdraw-address-1-fcoin.sql
btc-withdraw-address-2-fcoin.sql
wallet.dat
wallet.dat.meta
```

3. Clone BTC-Like address (Optional)

```
# clone deposit/exchange address
> ./bin/cloneaccount -s "./btc-deposit-addrs-1.txt" -t "./etp-deposit-address-1.sql" -q "./etp-ex-account-1.sql" -a "./etp-deposit-addrs-1.txt" -p "32" -i 142 -k "the-pubkey"
> ll
btc-deposit-addrs-1.txt
etp-deposit-address-1.sql
etp-deposit-addrs-1.txt
etp-ex-account-1.sql

# clone withdraw address sql
> ./bin/cloneaccount -s "./btc-withdraw-address-1.sql" -Q "./etp-withdraw-address-1.sql" -p "32"
> ll 
btc-withdraw-address-1.sql
etp-withdraw-address-1.sql
```

4. Convert address with same signer class (Optional)

```
# convert lsk address to algo address
> ./bin/cloneaccount -s "./lsk-withdraw-address-1.sql" -t "./algo-deposit-address-1.sql" -q "./algo-ex-account-1.sql" -a "./algo-deposit-addrs-1.txt" -Q "./algo-withdraw-address-1.sql" -i 172 -c "lsk->algo" -k "the-pubkey"
```

### Convert wallet to a new coin class with the same cryptography type and private key

```
# create a BTC wallet
> ./bin/walletctl -c BTC -n 1
> ll
btc-addrs.txt
wallet.dat

# convert to DOGE wallet
> ./bin/walletctl -c DOGE -g convert -o doge-wallet.dat
> ll
btc-addrs.txt
doge-addrs.txt
doge-wallet.dat
wallet.dat
```

### Convert wallet.dat of version v1 to version v2

```
> ./bin/walletctl --fromdatav1 -g convert -f lsk-wallet.dat -o lsk-wallet.v2.dat
> ll
lsk-addrs.txt
lsk-wallet.dat
lsk-wallet.v2.dat
```

### Convert wallet.dat of version pre-v1 to version v2

```
> ./bin/walletctl --fromdataprev1 -g convert -f lsk-wallet.dat -o lsk-wallet.v2.dat
> ll
lsk-addrs.txt
lsk-wallet.dat
lsk-wallet.v2.dat
```

### Convert wallet.dat of coinlib xlm to version v2

```
> ./bin/walletctl --fromdatacoinlibxlm -g convert -f xlm-wallet.dat -o xlm-wallet.v2.dat
> ll
xlm-addrs.txt
xlm-wallet.dat
xlm-wallet.v2.dat
```

### Convert seed.dat of iota-v1 to version v2

```
# -n = systemAddrNum + normalAddrNum
> ./bin/walletctl --fromdataiotav1 -g convert -f seed.dat -o iota-wallet.v2.dat -n 10
> ll
iota-addrs.txt
iota-wallet.v2.dat
seed.dat
```

### Convert wallet.dat of python-exwallet format to version v2

```
> ./bin/walletctl --fromdatapybtc ETP -g convert -f etp-wallet.dat -o etp-wallet.v2.dat
> ll
etp-addrs.txt
etp-wallet.dat
etp-wallet.v2.dat
```

### Profiling & Tracing

MacBook Pro: i7, 16G, SSD

Generate:
```
$ ./bin/genwallet -c ETH -n 5000000
Alloc: 1.3 GiB, HeapIdle: 38.7 MiB, HeapReleased: 0 B
INFO[2019-12-30 19:44:11.637] [generate] time cost: 2m18.497211654s         caller="log.go:162"
Alloc: 1.3 GiB, HeapIdle: 38.7 MiB, HeapReleased: 0 B
Alloc: 2.3 GiB, HeapIdle: 21.5 MiB, HeapReleased: 21.0 MiB
INFO[2019-12-30 19:47:45.158] [store] time cost: 3m33.521898126s            caller="log.go:162"
```

Rebuild:
```
$ ./bin/genwallet --rebuild        
Alloc: 1.3 GiB, HeapIdle: 24.0 MiB, HeapReleased: 0 B
INFO[2019-12-30 19:48:30.010] [load] time cost: 4.690481984s                caller="log.go:162"
Alloc: 1.3 GiB, HeapIdle: 24.0 MiB, HeapReleased: 0 B
Alloc: 2.3 GiB, HeapIdle: 44.8 MiB, HeapReleased: 43.8 MiB
INFO[2019-12-30 19:51:52.177] [store] time cost: 3m22.166155241s            caller="log.go:162"
```

## Use rundeck to deploy deposit service

```
deploy rundeck jobs for deposit.

Usage:
  rundeck [flags]

Examples:
./rundeck -u username -p password -i job-id -v v0.3.9.3 -s wallet-deposit-btc,wallet-deposit-eth

Flags:
  -c, --config string      the config file
  -h, --help               help for rundeck
  -i, --job string         deploy job id
  -p, --password string    the login password, for inputting in security mode, provide an empty string
  -s, --services strings   the services to deploy
  -u, --username string    the login username
  -v, --version string     the version to deploy
```

### Example of config

example-1:
```
; rundeck fcoin sandbox config.

(job "3db5fdce-91c3-416c-87e8-cba4cbeb02ce")

(username "username")

(version "v0.1")

(services
    "wallet-deposit-btc"
    "wallet-deposit-nodeapi")
```

example-2:
```
; rundeck fcoin sandbox config.

(job "3db5fdce-91c3-416c-87e8-cba4cbeb02ce")

(username "username")

(version "v0.1")

(services
    ; the same as (wallet-deposit-btc "v0.1")
    "wallet-deposit-btc"
    (wallet-deposit-nodeapi "nodeapi-v1")
    (wallet-deposit-tokens "tokens-v1"))
```

## Config of statuschecker

```
(emailpassword "123456")

(receivers
	"name@email.com")

(alives
	; (name host check-interval-minute)
	(eth1 "127.0.0.1:8080" 20))

(stucks
	; (name url check-interval-minute)
	(ADA "http://ip:port" 20))
```

## Use ecdsautil to convert secp256k1 public key in compressed and uncompressed format

```
> ./bin/ecdsautil -k 030ba7cddd228efef083735d27b632e4ae884fa2383b2a8c45d9da2b17f84a7481
compressed:
030ba7cddd228efef083735d27b632e4ae884fa2383b2a8c45d9da2b17f84a7481
uncompressed:
040ba7cddd228efef083735d27b632e4ae884fa2383b2a8c45d9da2b17f84a7481cc089c01874da96375e2fa83c2cde745311e2cfe931e1ed4ca0e4177d3580009
```

## Use genrsa generate rsa key pair

```
> ./bin/genrsa -f broadcast,signer,transfer
> ll
broadcast_rsa
broadcast_rsa.pub
signer_rsa
signer_rsa.pub
transfer_rsa
transfer_rsa.pub
```
