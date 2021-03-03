module github.com/solo-io/gloo-mesh-enterprise-helm

go 1.15

require (
	github.com/onsi/ginkgo v1.15.0
	github.com/onsi/gomega v1.10.5
	github.com/rotisserie/eris v0.5.0
	github.com/solo-io/gloo-mesh v1.0.0-beta6
	github.com/solo-io/gloo-mesh-enterprise v1.0.0-beta6
	github.com/solo-io/go-utils v0.20.5
	github.com/solo-io/skv2 v0.17.7
)

replace (
	// github.com/Azure/go-autorest/autorest has different versions for the Go
	// modules than it does for releases on the repository. Note the correct
	// version when updating.
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20200513103714-09dca8ec2884

	k8s.io/api => k8s.io/api v0.19.7
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.19.7
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.7
	k8s.io/client-go => k8s.io/client-go v0.19.7
	k8s.io/kubectl => k8s.io/kubectl v0.19.7
)
