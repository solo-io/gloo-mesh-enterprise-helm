package istio_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	extendertests "github.com/solo-io/gloo-mesh-enterprise/enterprise-extender/test/e2e/istio/pkg/tests"
	rbactests "github.com/solo-io/gloo-mesh-enterprise/rbac-webhook/test/e2e/pkg/tests"
	coretests "github.com/solo-io/gloo-mesh/test/e2e/istio/pkg/tests"
	"github.com/solo-io/gloo-mesh/test/utils"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/skv2/codegen/util"
)

var moduleRoot = util.GetModuleRoot()

// to skip testing this package, run `make run-tests SKIP_PACKAGES=test/e2e/istio
// to test only this package, run `make run-tests TEST_PKG=test/e2e/istio
func TestIstio(t *testing.T) {
	if os.Getenv("RUN_E2E") == "" {
		fmt.Println("skipping E2E tests")
		return
	}
	RegisterFailHandler(func(message string, callerSkip ...int) {
		utils.RunShell(util.GetModuleRoot() + "/ci/print-kind-info.sh")
		Fail(message, callerSkip...)
	})
	RunSpecs(t, "E2e Suite")
}

// Before running tests, federate the two clusters by creating a VirtualMesh with mTLS enabled.
var _ = BeforeSuite(func() {
	ensureWorkingDirectory()
	coretests.SetupClustersAndFederation(deployAndRegisterReleased)
})

func ensureWorkingDirectory() {
	// ensure we are in proper directory
	currentFile, err := testutils.GetCurrentFile()
	Expect(err).NotTo(HaveOccurred())
	currentDir := filepath.Dir(currentFile)
	utils.TestManifestDir = currentDir
	projectRoot := filepath.Join(currentDir, "..", "..", "..")
	err = os.Chdir(projectRoot)
	Expect(err).NotTo(HaveOccurred())
}

var _ = AfterSuite(func() {
	coretests.TeardownFederationAndClusters()
})

// initialize all tests
var _ = allTests()

func allTests() bool {
	/*
		test order matters here:
		- we run extender tests first because coretests modify settings (extender)
		- we run rbac tests last as rbactests modify rbac
	*/
	return Describe("enterprise helm chart", func() {
		extendertests.InitializeTests()
		coretests.InitializeTests()
		rbactests.InitializeTests()
	})
}

func deployAndRegisterReleased() {
	err := installEnterpriseChart()
	Expect(err).NotTo(HaveOccurred())
	err = registerCluster("mgmt-cluster")
	Expect(err).NotTo(HaveOccurred())
	err = registerCluster("remote-cluster")
	Expect(err).NotTo(HaveOccurred())

	// this sleep ensures rbac webhook certs are up to date before we start applying CRDs
	time.Sleep(time.Second * 10)
}

func installEnterpriseChart() error {
	chartPath := os.Getenv("OUTPUT_CHART_PATH")
	if chartPath == "" {
		if err := runCommand("make", "package-chart"); err != nil {
			return err
		}
		path, err := runCommandOut("make", "print-chart-path")
		if err != nil {
			return err
		}
		chartPath = path
	}
	licenseKey := os.Getenv("LICENSE_KEY")
	if chartPath == "" {
		return eris.Errorf("must provide var LICENSE_KEY with valid enterprise license key")
	}

	if err := runCommand(
		"kubectl",
		"create",
		"ns",
		"gloo-mesh",
	); err != nil && !strings.Contains(err.Error(), `namespaces "gloo-mesh" already exists`) {
		return err
	}

	if err := runCommand(
		"helm",
		"upgrade",
		"--install",
		"--namespace=gloo-mesh",
		"gloo-mesh-enterprise",
		chartPath,
		"--set",
		"licenseKey="+licenseKey,
		"--set",
		"gloo-mesh-ui.enabled=false", // disable apiserver/UI to free up compute for CI
	); err != nil {
		return err
	}

	for _, deployment := range []string{
		"discovery",
		"networking",
		"enterprise-extender",
		"rbac-webhook",
	} {
		if err := runCommand(
			"kubectl",
			"rollout",
			"status",
			"-n=gloo-mesh",
			"deployment",
			deployment,
		); err != nil {
			return err
		}
	}

	return nil
}

// uses meshctl from path to register cluster;
// make sure meshctl version matches gloo-mesh version from chart
func registerCluster(cluster string) error {
	apiServerAddr, err := getApiserverAddress(cluster)
	if err != nil {
		return err
	}
	if err := runCommand(
		"meshctl",
		"cluster",
		"register",
		"--mgmt-context=kind-mgmt-cluster",
		"--cluster-name="+cluster,
		"--remote-context=kind-"+cluster,
		"--api-server-address="+apiServerAddr,
	); err != nil {
		return err
	}

	return nil
}

func getApiserverAddress(cluster string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "host.docker.internal", nil
	default:
		addr, err := runCommandOut("bash", "-c", `docker exec `+cluster+`-control-plane ip addr show dev eth0 | sed -nE 's|\s*inet\s+([0-9.]+).*|\1|p'`)
		if err != nil {
			return "", err
		}
		return addr + ":6443", nil
	}
}

func runCommand(name string, args ...string) error {
	_, err := runCommandOut(name, args...)
	return err
}

// always runs command from module root
func runCommandOut(name string, args ...string) (string, error) {
	out := &bytes.Buffer{}

	cmd := exec.Command(
		name,
		args...,
	)
	cmd.Stdout = io.MultiWriter(out, GinkgoWriter)
	cmd.Stderr = io.MultiWriter(out, GinkgoWriter)
	cmd.Dir = moduleRoot

	fullCommand := append([]string{name}, args...)
	for _, arg := range args {
		if len(arg) > 64 {
			// silence license key in logs
			arg = arg[:64] + "..."
		}
	}

	log.Printf("running command %v", fullCommand)

	if err := cmd.Run(); err != nil {
		return "", eris.Wrapf(err, "running command (%v) failed: %v",
			fullCommand,
			out.String())
	}
	return out.String(), nil
}
