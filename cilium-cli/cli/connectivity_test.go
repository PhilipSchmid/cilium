// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cilium/cilium/cilium-cli/api"
	"github.com/cilium/cilium/cilium-cli/connectivity/check"
	"github.com/cilium/cilium/cilium-cli/defaults"
)

func TestNewConnectivityTests(t *testing.T) {
	testCases := []struct {
		params                            check.Parameters
		expectedCount                     int
		expectedTestNamespaces            []string
		expectedExternalTargetCANamespace []string
	}{
		{
			params: check.Parameters{
				FlowValidation:            check.FlowValidationModeDisabled,
				TestNamespace:             "cilium-test",
				ExternalTargetCANamespace: "",
			},
			expectedCount:                     1,
			expectedTestNamespaces:            []string{"cilium-test-1"},
			expectedExternalTargetCANamespace: []string{"cilium-test-1"},
		},
		{
			params: check.Parameters{
				FlowValidation:            check.FlowValidationModeDisabled,
				TestNamespace:             "cilium-test",
				ExternalTargetCANamespace: "cilium-test",
			},
			expectedCount:                     1,
			expectedTestNamespaces:            []string{"cilium-test-1"},
			expectedExternalTargetCANamespace: []string{"cilium-test"},
		},
		{
			params: check.Parameters{
				FlowValidation:            check.FlowValidationModeDisabled,
				TestNamespace:             "cilium-test",
				ExternalTargetCANamespace: "cilium-test",
				TestConcurrency:           -1,
			},
			expectedCount:                     1,
			expectedTestNamespaces:            []string{"cilium-test-1"},
			expectedExternalTargetCANamespace: []string{"cilium-test"},
		},
		{
			params: check.Parameters{
				FlowValidation:            check.FlowValidationModeDisabled,
				TestNamespace:             "cilium-test",
				ExternalTargetCANamespace: "",
				TestConcurrency:           3,
			},
			expectedCount:                     3,
			expectedTestNamespaces:            []string{"cilium-test-1", "cilium-test-2", "cilium-test-3"},
			expectedExternalTargetCANamespace: []string{"cilium-test-1", "cilium-test-2", "cilium-test-3"},
		},
		{
			params: check.Parameters{
				FlowValidation:            check.FlowValidationModeDisabled,
				TestNamespace:             "cilium-test",
				ExternalTargetCANamespace: "cilium-test",
				TestConcurrency:           3,
			},
			expectedCount:                     3,
			expectedTestNamespaces:            []string{"cilium-test-1", "cilium-test-2", "cilium-test-3"},
			expectedExternalTargetCANamespace: []string{"cilium-test"},
		},
	}
	for _, tt := range testCases {
		// function to test
		actual, err := newConnectivityTests(tt.params, &api.NopHooks{}, check.NewConcurrentLogger(&bytes.Buffer{}, 1))

		require.NoError(t, err)
		require.Len(t, actual, tt.expectedCount)
		for i, n := range tt.expectedTestNamespaces {
			require.Equal(t, n, actual[i].Params().TestNamespace)
		}
		for i, n := range tt.expectedExternalTargetCANamespace {
			require.Equal(t, n, actual[i].Params().ExternalTargetCANamespace)
		}
	}
}

func TestConnectivityTestFlags(t *testing.T) {
	ct := newCmdConnectivityTest(&api.NopHooks{})
	require.Empty(t, params.JunitProperties)
	ct.Flags().Set("junit-property", "a=b")
	require.NoError(t, ct.Flags().Set("junit-property", "a=b"))
	require.Equal(t, map[string]string{"a": "b"}, params.JunitProperties)
	require.NoError(t, ct.Flags().Set("junit-property", "c=d"))
	require.Equal(t, map[string]string{"a": "b", "c": "d"}, params.JunitProperties)

	require.Empty(t, params.NodeSelector)
	require.NoError(t, ct.Flags().Set("node-selector", "a=b"))
	require.Equal(t, map[string]string{"a": "b"}, params.NodeSelector)
	require.NoError(t, ct.Flags().Set("node-selector", "c=d"))
	require.Equal(t, map[string]string{"a": "b", "c": "d"}, params.NodeSelector)
}

func TestPrintImageArtifacts(t *testing.T) {
	ct := newCmdConnectivityTest(&api.NopHooks{})
	var buf bytes.Buffer

	// Test print-image-artifacts flag for connectivity test subcommand
	buf.Reset()
	ct.SetOut(&buf)
	require.NoError(t, ct.Flags().Set("print-image-artifacts", "true"))
	require.NoError(t, ct.Execute())
	expectedImages := []string{
		defaults.ConnectivityCheckAlpineCurlImage,
		defaults.ConnectivityCheckJSONMockImage,
		defaults.ConnectivityDNSTestServerImage,
		defaults.ConnectivityTestConnDisruptImage,
		defaults.ConnectivityTestFRRImage,
		defaults.ConnectivityTestSocatImage,
	}
	for _, img := range expectedImages {
		require.Contains(t, buf.String(), img)
	}

	// Test print-image-artifacts flag for connectivity test subcommand with overridden image
	buf.Reset()
	ct.SetOut(&buf)
	var alpineImage = "alpine/curl:latest"
	require.NoError(t, ct.Flags().Set("print-image-artifacts", "true"))
	require.NoError(t, ct.Flags().Set("curl-image", alpineImage))
	require.NoError(t, ct.Execute())
	expectedImagesWithOverride := []string{
		alpineImage,
		defaults.ConnectivityCheckJSONMockImage,
		defaults.ConnectivityDNSTestServerImage,
		defaults.ConnectivityTestConnDisruptImage,
		defaults.ConnectivityTestFRRImage,
		defaults.ConnectivityTestSocatImage,
	}
	for _, img := range expectedImagesWithOverride {
		require.Contains(t, buf.String(), img)
	}

	// Ensure no images are printed when the flag is not set
	buf.Reset()
	ct.SetOut(&buf)
	require.NoError(t, ct.Flags().Set("print-image-artifacts", "false"))
	require.NoError(t, ct.Execute())
	for _, img := range expectedImages {
		require.NotContains(t, buf.String(), img)
		require.Empty(t, buf.String())
	}

	// Test print-image-artifacts flag for connectivity perf subcommand
	cp := newCmdConnectivityPerf(&api.NopHooks{})
	buf.Reset()
	cp.SetOut(&buf)
	require.NoError(t, cp.Flags().Set("print-image-artifacts", "true"))
	require.NoError(t, cp.Execute())
	require.Contains(t, buf.String(), defaults.ConnectivityPerformanceImage)

	// Test print-image-artifacts flag for connectivity perf subcommand with overridden image
	buf.Reset()
	cp.SetOut(&buf)
	var perfImage = "alpine:latest"
	require.NoError(t, cp.Flags().Set("print-image-artifacts", "true"))
	require.NoError(t, ct.Flags().Set("performance-image", perfImage))
	require.NoError(t, cp.Execute())
	require.Contains(t, buf.String(), perfImage)
	require.NotContains(t, buf.String(), defaults.ConnectivityPerformanceImage)

	// Ensure no images are printed when the flag is not set
	buf.Reset()
	cp.SetOut(&buf)
	require.NoError(t, cp.Flags().Set("print-image-artifacts", "false"))
	require.NoError(t, cp.Execute())
	require.NotContains(t, buf.String(), defaults.ConnectivityPerformanceImage)
}
