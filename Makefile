.PHONY: test-acc
test-acc:
	TF_ACC=1 \
	go test ./provider/ -v -run=^TestAcc_ -count=1 -coverprofile=cov.out -coverpkg "github.com/ctfer-io/terraform-provider-ctfd/v2/provider,github.com/ctfer-io/terraform-provider-ctfd/v2/provider/utils,github.com/ctfer-io/terraform-provider-ctfd/v2/provider/validators"

.PHONY: docs
docs:
	go generate ./...
