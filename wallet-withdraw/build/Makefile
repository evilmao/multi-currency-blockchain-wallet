MK := ./make/make.sh
MK_GIT_VERSION := ./make/make_gitversion.sh
MK_IMPORT := ./make/make_import.sh
MK_IMPORT_CLAIM := ./make/make_import_claim.sh

all: signer transfer broadcast syncd

libmonero:
ifeq ("${buildxmr}","1")
	./make/build_monero.sh
endif

version:
	$(MK_GIT_VERSION) ../cmd cmd

signer: version libmonero
	$(MK) signer signer

transfer: version libmonero
	$(MK_IMPORT) ../transfer/txbuilder upex-wallet/wallet-withdraw/transfer/txbuilder  ../cmd/transfer imports
	${MK} transfer transfer

broadcast: version
	$(MK_IMPORT) ../broadcast/handler upex-wallet/wallet-withdraw/broadcast/handler ../cmd/broadcast imports
	$(MK) broadcast broadcast

syncd: version
	$(MK) syncd syncd

.PHONY: all version libmonero signer transfer broadcast syncd
