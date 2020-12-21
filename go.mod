module github.com/solo-io/gloo-mesh-enterprise-helm

go 1.15

require (
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	github.com/rotisserie/eris v0.4.0
	github.com/solo-io/gloo-mesh v0.11.2
	github.com/solo-io/gloo-mesh-enterprise v0.3.4
	github.com/solo-io/go-utils v0.20.0
	github.com/solo-io/skv2 v0.15.2
)

replace (
	// github.com/Azure/go-autorest/autorest has different versions for the Go
	// modules than it does for releases on the repository. Note the correct
	// version when updating.
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.2
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

	// needed until github.com/solo-io/wasm/tools/wasme/pkg
	github.com/solo-io/go-utils => github.com/solo-io/go-utils v0.19.0

	// https://github.com/ory/dockertest/issues/208#issuecomment-686820414
	golang.org/x/sys => golang.org/x/sys v0.0.0-20200826173525-f9321e4c35a6

	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20200513103714-09dca8ec2884

	k8s.io/client-go => k8s.io/client-go v0.18.6

)
