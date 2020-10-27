MK := ./make.sh

all: statuschecker cloneaccount genwallet walletctl rundeck ecdsautil genrsa cmdserver

libmonero:
ifeq ("${buildxmr}","1")
	./build_monero.sh
endif

statuschecker:
	$(MK) statuschecker statuschecker

cloneaccount: libmonero
	$(MK) cloneaccount cloneaccount

genwallet: libmonero
	$(MK) genwallet/main/genwallet genwallet

walletctl: libmonero
	$(MK) genwallet/main/walletctl walletctl

rundeck:
	$(MK) rundeck rundeck

ecdsautil:
	$(MK) ecdsautil ecdsautil

genrsa:
	$(MK) genrsa genrsa

cmdserver:
	$(MK) cmdserver cmdserver

.PHONY: all statuschecker cloneaccount genwallet walletctl rundeck ecdsautil genrsa cmdserver
