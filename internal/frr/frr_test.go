// SPDX-License-Identifier:Apache-2.0

package frr

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	v1alpha1 "github.com/metallb/frrk8s/api/v1alpha1"
	"github.com/metallb/frrk8s/internal/logging"
	"k8s.io/apimachinery/pkg/util/wait"
)

const testData = "testdata/"

var classCMask = net.IPv4Mask(0xff, 0xff, 0xff, 0)

var update = flag.Bool("update", false, "update .golden files")

func testOsHostname() (string, error) {
	return "dummyhostname", nil
}

func testCompareFiles(t *testing.T, configFile, goldenFile string) {
	var lastError error

	// Try comparing files multiple times because tests can generate more than one configuration
	err := wait.PollImmediate(10*time.Millisecond, 2*time.Second, func() (bool, error) {
		lastError = nil
		cmd := exec.Command("diff", configFile, goldenFile)
		output, err := cmd.Output()

		if err != nil {
			lastError = fmt.Errorf("command %s returned error: %s\n%s", cmd.String(), err, output)
			return false, nil
		}

		return true, nil
	})

	// err can only be a ErrWaitTimeout, as the check function always return nil errors.
	// So lastError is always set
	if err != nil {
		t.Fatalf("failed to compare configfiles %s, %s using poll interval\nlast error: %v", configFile, goldenFile, lastError)
	}
}

func testUpdateGoldenFile(t *testing.T, configFile, goldenFile string) {
	t.Log("update golden file")

	// Sleep to be sure the sessionManager has produced all configuration the test
	// has triggered and no config is still waiting in the debouncer() local variables.
	// No other conditions can be checked, so sleeping is our best option.
	time.Sleep(100 * time.Millisecond)

	cmd := exec.Command("cp", "-a", configFile, goldenFile)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("command %s returned %s and error: %s", cmd.String(), output, err)
	}
}

func testGenerateFileNames(t *testing.T) (string, string) {
	return filepath.Join(testData, filepath.FromSlash(t.Name())), filepath.Join(testData, filepath.FromSlash(t.Name())+".golden")
}

func testSetup(t *testing.T) {
	configFile, _ := testGenerateFileNames(t)
	os.Setenv("FRR_CONFIG_FILE", configFile)
	_ = os.Remove(configFile) // removing leftovers from previous runs
	osHostname = testOsHostname
}

func testCheckConfigFile(t *testing.T) {
	configFile, goldenFile := testGenerateFileNames(t)

	if *update {
		testUpdateGoldenFile(t, configFile, goldenFile)
	}

	testCompareFiles(t, configFile, goldenFile)

	if !strings.Contains(configFile, "Invalid") {
		err := testFileIsValid(configFile)
		if err != nil {
			t.Fatalf("Failed to verify the file %s", err)
		}
	}
}

func TestSingleSession(t *testing.T) {
	testSetup(t)
	ctx, cancel := context.WithCancel(context.Background())
	frr := NewFRR(ctx, logging.LevelInfo)
	defer cancel()

	config := v1alpha1.FRRConfiguration{
		Spec: v1alpha1.FRRConfigurationSpec{
			Routers: []v1alpha1.Router{
				{
					ASN: 65000,
					Neighbors: []v1alpha1.Neighbor{
						{
							ASN:      65001,
							Address:  "192.168.1.2",
							Port:     4567,
							Password: "password",
							AllowedOutPrefixes: v1alpha1.AllowedPrefixes{
								Prefixes: []string{
									"192.169.1.0/24",
									"192.170.1.0/22",
								},
							},
						},
					},
				},
			},
		},
	}
	err := frr.ApplyConfig(config)
	if err != nil {
		t.Fatalf("Failed to apply config: %s", err)
	}

	testCheckConfigFile(t)
}
