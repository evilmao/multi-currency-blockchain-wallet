
MK := ./make/make.sh
MK_GIT_VERSION := ./make/make_gitversion.sh
MK_DEPOSIT := ./make/make_deposit.sh
MK_IMPORT := ./make/make_import.sh
ServicesFile := ../.circleci/all_services.sh

all:
	#@$(MAKE) prebuild
	@$(MAKE) build
	@$(MAKE) postbuild

version:
	$(MK_GIT_VERSION) ../cmd cmd

deposit: version
	$(MK_IMPORT) ../rpc upex-wallet/wallet-deposit/rpc ../cmd imports
	@go install ../cmd/deposit
	go build -o ./bin/deposit ../cmd/deposit
	@echo

build: deposit
	@echo "declare -a all_services=(" > ${ServicesFile}

postbuild:
	@echo ")" >> ${ServicesFile}

.PHONY: all version deposit build postbuild
