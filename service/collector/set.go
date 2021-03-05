package collector

import (
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/certs/v3/pkg/certs"
	"github.com/giantswarm/exporterkit/collector"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type SetConfig struct {
	CertSearcher certs.Interface
	K8sClient    k8sclient.Interface
	Logger       micrologger.Logger

	NewCommonClusterObjectFunc func() infrastructurev1alpha2.CommonClusterObject
}

// Set is basically only a wrapper for the operator's collector implementations.
// It eases the initialization and prevents some weird import mess so we do not
// have to alias packages.
type Set struct {
	*collector.Set
}

func NewSet(config SetConfig) (*Set, error) {
	var err error

	var collectorSet *collector.Set
	{
		c := collector.SetConfig{
			Collectors: []collector.Interface{},
			Logger:     config.Logger,
		}

		collectorSet, err = collector.NewSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Set{
		Set: collectorSet,
	}

	return s, nil
}
