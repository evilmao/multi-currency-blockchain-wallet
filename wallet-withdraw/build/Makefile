MK := ./make/make.sh
MK_GITVERSION := ./make/make_gitversion.sh
MK_IMPORT := ./make/make_import.sh
MK_IMPORT_CLAIM := ./make/make_import_claim.sh

all: signer transfer broadcast

libmonero:
ifeq ("${buildxmr}","1")
	./make/build_monero.sh
endif

version:
	$(MK_GITVERSION) ../cmd cmd

signer: version libmonero
	$(MK) signer signer

transfer: version libmonero
	$(MK_IMPORT) ../transfer/txbuilder upex-wallet/wallet-withdraw/transfer/txbuilder  ../cmd/transfer main
	$(MK_IMPORT_CLAIM) ../transfer/claimer github.com/fb996de/wallet-withdraw/transfer/claimer ../cmd/transfer main
	${MK} transfer transfer

broadcast: version
	$(MK_IMPORT) ../broadcast/handler upex-wallet/wallet-withdraw/broadcast/handler ../cmd/broadcast main
	$(MK) broadcast broadcast

broadcast: version
	$(MK_IMPORT) ../broadcast/handler github.com/fb996de/wallet-withdraw/broadcast/handler ../cmd/broadcast main
	$(MK) broadcast broadcast

.PHONY: all version libmonero signer transfer broadcast
