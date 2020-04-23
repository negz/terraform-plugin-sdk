module github.com/hashicorp/terraform-plugin-sdk/v2

go 1.14

replace github.com/hashicorp/terraform-plugin-test v1.2.0 => ../terraform-plugin-test

require (
	github.com/apparentlymart/go-cidr v1.0.1
	github.com/apparentlymart/go-dump v0.0.0-20190214190832-042adf3cf4a0
	github.com/davecgh/go-spew v1.1.1
	github.com/go-test/deep v1.0.3
	github.com/golang/mock v1.3.1
	github.com/golang/protobuf v1.3.4
	github.com/golang/snappy v0.0.1
	github.com/google/go-cmp v0.3.1
	github.com/hashicorp/errwrap v1.0.0
	github.com/hashicorp/go-cleanhttp v0.5.1
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/go-plugin v1.2.2
	github.com/hashicorp/go-uuid v1.0.1
	github.com/hashicorp/go-version v1.2.0
	github.com/hashicorp/hcl/v2 v2.0.0
	github.com/hashicorp/logutils v1.0.0
	github.com/hashicorp/terraform-json v0.4.0
	github.com/hashicorp/terraform-plugin-sdk v1.11.0
	github.com/hashicorp/terraform-plugin-test v1.3.0
	github.com/keybase/go-crypto v0.0.0-20161004153544-93f5b35093ba
	github.com/mitchellh/copystructure v1.0.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-testing-interface v1.0.4
	github.com/mitchellh/mapstructure v1.1.2
	github.com/mitchellh/reflectwalk v1.0.1
	github.com/pierrec/lz4 v2.0.5+incompatible
	github.com/zclconf/go-cty v1.2.1
	golang.org/x/crypto v0.0.0-20190820162420-60c769a6c586
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a
	golang.org/x/tools v0.0.0-20190628153133-6cdbf07be9d0
	google.golang.org/grpc v1.27.1
)
