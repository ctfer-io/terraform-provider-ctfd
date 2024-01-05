.PHONY: test-acc
test-acc:
	TF_ACC=1 \
	go test ./... -v -run=^TestAcc_ -count=1 -coverprofile=cov.out -coverpkg "github.com/ctfer-io/terraform-provider-ctfd/internal/provider,github.com/ctfer-io/terraform-provider-ctfd/internal/provider/challenge,github.com/ctfer-io/terraform-provider-ctfd/internal/provider/utils,github.com/ctfer-io/terraform-provider-ctfd/internal/provider/validators"
