package process

import (
	"testing"

	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/kyma-project/kyma-metrics-collector/env"
	"github.com/kyma-project/kyma-metrics-collector/pkg/edp"
	skrredis "github.com/kyma-project/kyma-metrics-collector/pkg/skr/redis"
	kmctesting "github.com/kyma-project/kyma-metrics-collector/pkg/testing"
)

func TestParse(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	config := &env.Config{PublicCloudSpecsPath: testPublicCloudSpecsPath}
	publicCloudSpecs, err := LoadPublicCloudSpecs(config)
	g.Expect(err).Should(gomega.BeNil())

	testCases := []struct {
		name            string
		input           Input
		specs           PublicCloudSpecs
		expectedMetrics edp.ConsumptionMetrics
		expectedErr     bool
	}{
		{
			name: "with Azure, 2 vm types, 1 NFS pvc (20Gi) and 2 svcs(1 clusterIP and 1 LoadBalancer)",
			input: Input{
				provider: Azure,
				nodeList: kmctesting.Get2Nodes(),
				pvcList:  kmctesting.Get1NFSPVC(),
				svcList:  kmctesting.Get2SvcsOfDiffTypes(),
			},
			specs: *publicCloudSpecs,
			expectedMetrics: edp.ConsumptionMetrics{
				// ResourceGroups: nil,
				Compute: edp.Compute{
					VMTypes: []edp.VMType{{
						Name:  "standard_d8_v3",
						Count: 2,
					}},
					ProvisionedCpus:  16,
					ProvisionedRAMGb: 64,
					ProvisionedVolumes: edp.ProvisionedVolumes{
						SizeGbTotal:   60,
						Count:         1,
						SizeGbRounded: 64,
					},
				},
			},
		},
		{
			name: "with Azure, 2 vm types, 3 pvcs(5,10 and 20Gi) and 2 svcs(1 clusterIP and 1 LoadBalancer)",
			input: Input{
				provider: Azure,
				nodeList: kmctesting.Get2Nodes(),
				pvcList:  kmctesting.Get3PVCs(),
				svcList:  kmctesting.Get2SvcsOfDiffTypes(),
			},
			specs: *publicCloudSpecs,
			expectedMetrics: edp.ConsumptionMetrics{
				// ResourceGroups: nil,
				Compute: edp.Compute{
					VMTypes: []edp.VMType{{
						Name:  "standard_d8_v3",
						Count: 2,
					}},
					ProvisionedCpus:  16,
					ProvisionedRAMGb: 64,
					ProvisionedVolumes: edp.ProvisionedVolumes{
						SizeGbTotal:   35,
						Count:         3,
						SizeGbRounded: 96,
					},
				},
			},
		},
		{
			name: "with Azure with 3 vms and no pvc and svc",
			input: Input{
				provider: Azure,
				nodeList: kmctesting.Get3NodesWithStandardD8v3VMType(),
			},
			specs: *publicCloudSpecs,
			expectedMetrics: edp.ConsumptionMetrics{
				// ResourceGroups: nil,
				Compute: edp.Compute{
					VMTypes: []edp.VMType{{
						Name:  "standard_d8_v3",
						Count: 3,
					}},
					ProvisionedCpus:  24,
					ProvisionedRAMGb: 96,
					ProvisionedVolumes: edp.ProvisionedVolumes{
						SizeGbTotal:   0,
						Count:         0,
						SizeGbRounded: 0,
					},
				},
			},
		},
		{
			name: "with Azure with 3 vms and no pvc and svc",
			input: Input{
				provider: Azure,
				nodeList: kmctesting.Get3NodesWithStandardD8v3VMType(),
			},
			specs: *publicCloudSpecs,
			expectedMetrics: edp.ConsumptionMetrics{
				Compute: edp.Compute{
					VMTypes: []edp.VMType{{
						Name:  "standard_d8_v3",
						Count: 3,
					}},
					ProvisionedCpus:  24,
					ProvisionedRAMGb: 96,
					ProvisionedVolumes: edp.ProvisionedVolumes{
						SizeGbTotal:   0,
						Count:         0,
						SizeGbRounded: 0,
					},
				},
			},
		},
		{
			name: "with Azure with 3 vms and 2 redis, no pvc and svc",
			input: Input{
				provider: Azure,
				nodeList: kmctesting.Get3NodesWithStandardD8v3VMType(),
				redisList: &skrredis.RedisList{
					Azure: *kmctesting.AzureRedisList(),
				},
			},
			specs: *publicCloudSpecs,
			expectedMetrics: edp.ConsumptionMetrics{
				Compute: edp.Compute{
					VMTypes: []edp.VMType{{
						Name:  "standard_d8_v3",
						Count: 3,
					}},
					ProvisionedCpus:  24,
					ProvisionedRAMGb: 96,
					ProvisionedVolumes: edp.ProvisionedVolumes{
						SizeGbTotal:   5731,
						Count:         2,
						SizeGbRounded: 5731,
					},
				},
			},
		},
		{
			name: "with Azure and vm type missing from the list of vmtypes",
			input: Input{
				provider: Azure,
				nodeList: kmctesting.Get3NodesWithFooVMType(),
			},
			specs:       *publicCloudSpecs,
			expectedErr: true,
		},
		{
			name: "with sapconvergedcloud, 2 vm types, 3 pvcs(5,10 and 20Gi), and 2 svcs(1 clusterIP and 1 LoadBalancer)",
			input: Input{
				provider: CCEE,
				nodeList: kmctesting.Get2NodesOpenStack(),
				pvcList:  kmctesting.Get3PVCs(),
				svcList:  kmctesting.Get2SvcsOfDiffTypes(),
			},
			specs: *publicCloudSpecs,
			expectedMetrics: edp.ConsumptionMetrics{
				Compute: edp.Compute{
					VMTypes: []edp.VMType{{
						Name:  "g_c12_m48",
						Count: 2,
					}},
					ProvisionedCpus:  24,
					ProvisionedRAMGb: 96,
					ProvisionedVolumes: edp.ProvisionedVolumes{
						SizeGbTotal:   35,
						Count:         3,
						SizeGbRounded: 96,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotMetrics, err := tc.input.Parse(&tc.specs)
			if !tc.expectedErr {
				g.Expect(err).Should(gomega.BeNil())
				g.Expect(gotMetrics.Compute).To(gomega.Equal(tc.expectedMetrics.Compute))
				g.Expect(gotMetrics.Networking).To(gomega.Equal(tc.expectedMetrics.Networking))
				g.Expect(gotMetrics.Timestamp).To(gomega.Not(gomega.BeEmpty()))

				return
			}

			g.Expect(err).ShouldNot(gomega.BeNil())
			g.Expect(gotMetrics).Should(gomega.BeNil())
		})
	}
}

func TestGetSizeInGB(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	testCases := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name:     "value in GB",
			input:    "15Gi",
			expected: 15,
		},
		{
			name:     "value in GB again",
			input:    "10Gi",
			expected: 10,
		},
		{
			name:     "value in TB",
			input:    "10Ti",
			expected: 10240,
		},
		{
			name:     "value in MB",
			input:    "10Mi",
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := resource.MustParse(tc.input)
			got := getSizeInGB(&input)
			g.Expect(tc.expected).To(gomega.Equal(got))
		})
	}
}

func TestGetVolumeRoundedToFactor(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	testCases := []struct {
		input    int64
		expected int64
	}{
		{
			input:    31,
			expected: 32,
		},
		{
			input:    42,
			expected: 64,
		},
		{
			input:    64,
			expected: 64,
		},
		{
			input:    32,
			expected: 32,
		},
		{
			input:    0,
			expected: 0,
		},
		{
			input:    654,
			expected: 672,
		},
	}

	for _, tc := range testCases {
		got := getVolumeRoundedToFactor(tc.input)
		g.Expect(tc.expected).To(gomega.Equal(got))
	}
}
