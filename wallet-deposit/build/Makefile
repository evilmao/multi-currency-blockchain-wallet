
MK := ./make/make.sh
MK_GIT_VERSION := ./make/make_gitversion.sh
MK_DEPOSIT := ./make/make_deposit.sh
MK_IMPORT := ./make/make_import.sh
ServicesFile := ../.circleci/all_services.sh

all:
	@$(MAKE) build

version:
	$(MK_GIT_VERSION) ../cmd cmd

prebuild: version
	$(MK_IMPORT) ../rpc upex-wallet/wallet-deposit/rpc ../cmd imports

build: prebuild
	@go install ../cmd/deposit
	go build -o ./bin/deposit ../cmd/deposit
	@echo


.PHONY: all version prebuild build
