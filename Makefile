.DEFAULT: all

all: golint_by_docker golangci_lint_by_docker build

protos: generate_proto

generate_proto:
	sh scripts/generate_proto.sh  "go" "./" "./go"

generate_deploy:
	sh scripts/generate_deploys.sh

golangci_lint:
	bash scripts/golangci_lint.sh

golangci_lint_by_docker:
	bash scripts/golangci_lint_by_docker.sh

golint_by_docker:
	bash scripts/golint_by_docker.sh

build:
	bash go_build.sh
