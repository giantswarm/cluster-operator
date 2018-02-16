package service

import (
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/spf13/viper"

	"github.com/giantswarm/cluster-operator/flag"
)

func Test_Service_New(t *testing.T) {
	testCases := []struct {
		description          string
		config               func() Config
		expectedErrorHandler func(error) bool
	}{
		{
			description: "empty value config must return invalidConfigError",
			config: func() Config {
				return Config{}
			},
			expectedErrorHandler: IsInvalidConfig,
		},
		{
			description: "production-like config must be valid",
			config: func() Config {
				config := Config{}

				config.Logger = microloggertest.New()

				config.Flag = flag.New()
				config.Viper = viper.New()

				config.Description = "test"
				config.GitCommit = "test"
				config.ProjectName = "test"
				config.Source = "test"

				config.Viper.Set(config.Flag.Service.Kubernetes.Address, "http://127.0.0.1:6443")
				config.Viper.Set(config.Flag.Service.Kubernetes.InCluster, "false")

				return config
			},
			expectedErrorHandler: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			_, err := New(tc.config())

			if err != nil {
				if tc.expectedErrorHandler == nil {
					t.Fatalf("unexpected error returned: %#v", err)
				}
				if !tc.expectedErrorHandler(err) {
					t.Fatalf("incorrect error returned: %#v", err)
				}
			}
		})
	}
}
