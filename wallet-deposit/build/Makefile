
MK := ./make/make.sh
MK_GITVERSION := ./make/make_gitversion.sh
MK_DEPOSIT := ./make/make_deposit.sh
MK_IMPORT := ./make/make_import.sh
ServicesFile := ../.circleci/all_services.sh

all:
	@$(MAKE) prebuild
	@$(MAKE) build
	@$(MAKE) postbuild

version:
	$(MK_GITVERSION) ../cmd cmd

deposit: version
	$(MK_IMPORT) ../rpc upex-wallet/wallet-deposit/rpc ../cmd/deposit main
	@go install ../cmd/deposit
	go build -o ./bin/deposit ../cmd/deposit
	@echo

prebuild: deposit
	@echo "declare -a all_services=(" > ${ServicesFile}

build: eth

postbuild:
	@echo ")" >> ${ServicesFile}

include ./make/*.mk
.PHONY: all version deposit prebuild build postbuild
