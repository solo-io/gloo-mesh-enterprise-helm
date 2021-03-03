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

	"github.com/solo-io/gloo-mesh-enterprise-helm/test/e2e/istio/pkg"
	"github.com/solo-io/gloo-mesh/test/extensions"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
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
		log.Printf("Failed with message: %v", message)
		utils.RunShell(util.GetModuleRoot() + "/ci/print-kind-info.sh")
		Fail(message, callerSkip...)
	})
	RunSpecs(t, "E2e Suite")
}

// Before running tests, federate the two clusters by creating a VirtualMesh with mTLS enabled.
var _ = BeforeSuite(func() {
	ensureWorkingDirectory()
	// deployAndRegisterEnterprise()
	// coretests.SetupClustersAndFederation(deployAndRegisterEnterprise)
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
		// extendertests.InitializeTests()
		// coretests.InitializeTests()
		// rbactests.InitializeTests()
	})
}

func deployAndRegisterEnterprise() {
	// override Docker Host Addr for CircleCI
	extensions.DockerHostAddress = pkg.DockerHostAddress

	err := installEnterpriseChart()
	Expect(err).NotTo(HaveOccurred())
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
		"enterprise-networking",
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

func getApiserverAddress(cluster string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "host.docker.internal", nil
	default:
		addr, err := runCommandOut("bash", "-c", `docker exec `+cluster+`-control-plane ip addr show dev eth0 | sed -nE 's|\s*inet\s+([0-9.]+).*|\1|p'`)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(addr) + ":6443", nil
	}
}

func applyCustomBootstrapPatch(clusterName, namespace, deploymentName string) error {
	if _, err := runCommandInOut(
		"kubectl",
		pkg.CustomBootsrapConfigmap(namespace),
		"--context=kind-"+clusterName,
		"apply",
		"-f-",
	); err != nil {
		return err
	}

	if err := runCommand(
		"kubectl",
		"patch",
		"deployment",
		"-n", namespace,
		deploymentName,
		"--context=kind-"+clusterName,
		fmt.Sprintf("--patch=%v", pkg.CustomBootstrapOverridePatch),
	); err != nil {
		return err
	}

	return nil
}

func runCommand(name string, args ...string) error {
	_, err := runCommandOut(name, args...)
	return err
}

// always runs command from module root
func runCommandOut(name string, args ...string) (string, error) {
	return runCommandInOut(name, "", args...)
}

// always runs command from module root
func runCommandInOut(name, stdin string, args ...string) (string, error) {
	out := &bytes.Buffer{}

	cmd := exec.Command(
		name,
		args...,
	)
	if stdin != "" {
		cmd.Stdin = bytes.NewBuffer([]byte(stdin))
	}
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
