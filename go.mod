module kubevirt.io/vhostuser-network-binding-plugin

go 1.26.3

require (
	github.com/onsi/ginkgo/v2 v2.29.0
	github.com/onsi/gomega v1.41.0
	google.golang.org/grpc v1.81.1
	k8s.io/api v0.36.1
	k8s.io/dynamic-resource-allocation v0.36.1
	k8s.io/klog/v2 v2.140.0
	kubevirt.io/api v1.9.0-beta.0.0.20260608175919-2d80cea6b069
	kubevirt.io/client-go v1.9.0-beta.0.0.20260608175921-88352ed030c2
	kubevirt.io/kubevirt v1.9.0-beta.0.0.20260608173016-22f8e03db1be
	libvirt.org/go/libvirtxml v1.12002.0
)

require (
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20260402051712-545e8a4df936 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/openshift/custom-resource-status v1.1.2 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.uber.org/mock v0.5.1 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/mod v0.35.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	golang.org/x/tools v0.44.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260226221140-a57be14db171 // indirect
	google.golang.org/protobuf v1.36.12-0.20260120151049-f2248ac996af // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/apiextensions-apiserver v0.34.3 // indirect
	k8s.io/apimachinery v0.36.1 // indirect
	k8s.io/utils v0.0.0-20260210185600-b8788abfbbc2 // indirect
	kubevirt.io/containerized-data-importer-api v1.64.0 // indirect
	kubevirt.io/controller-lifecycle-operator-sdk/api v0.2.4 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.2 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)

// k8s.io/dynamic-resource-allocation v0.36.1 is required for the DRA downward
// API metadata packages and depends on k8s/apimachinery v0.36.1. Kubevirt pins
// k8s.io/apimachinery to v0.34.x so adding the DRA apis breaks the dependency tree.
//
// The replace directives below pin the k8s.io/* modules back to v0.34.3.
// Luckily, the DRA v0.36.1 metadata packages compile cleanly against k8s.io/apimachinery
// v0.34.3.
replace (
	k8s.io/api => k8s.io/api v0.34.3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.34.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.34.3
)
