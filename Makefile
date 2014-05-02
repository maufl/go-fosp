all: vet lint fmt install

vet:
	@echo "Vetting.."
	@-go vet fosp/*.go
	@-go vet fospc/*.go
	@-go vet fospd/*.go
	@-go vet fbts/*.go
	@echo

lint:
	@echo "Linting.."
	@golint */*.go
	@echo

fmt:
	@echo "Fmting.."
	@go fmt fosp/*.go
	@go fmt fospc/*.go
	@go fmt fospd/*.go
	@go fmt fbts/*.go
	@echo

install:
	@echo "Installing.."
	@(cd fosp && go install)
	@(cd fospc && go install)
	@(cd fospd && go install)
	@(cd fbts && go install)
	@echo
