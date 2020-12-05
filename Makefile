.DEFAULT: golint_by_docker golangci_lint_by_docker

golangci_lint:
	bash scripts/golangci_lint.sh

golangci_lint_by_docker:
	bash scripts/golangci_lint_by_docker.sh

golint_by_docker:
	bash scripts/golint_by_docker.sh
