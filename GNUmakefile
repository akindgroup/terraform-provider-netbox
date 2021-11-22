GOOS=$(shell uname | awk '{print tolower($$0)}')
GOARCH=amd64
TERRAFORM_PLUGIN_DIR=$(HOME)/.terraform.d/plugins/academicwork/netbox/0.1.1/${GOOS}_${GOARCH}
default: testacc
.PHONY: testacc build_dev
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m
build_dev:
	mkdir -p ${TERRAFORM_PLUGIN_DIR}
	go build -o ${TERRAFORM_PLUGIN_DIR}/terraform-provider-netbox
