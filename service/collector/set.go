package collector

import (
	"github.com/giantswarm/exporterkit/collector"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type SetConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

// Set is basically only a wrapper for the operator's collector implementations.
// It eases the initialization and prevents some weird import mess so we do not
// have to alias packages. There is also the benefit of the helper type kept
// private so we do not need to expose this magic.
type Set struct {
	*collector.Set
}

func NewSet(config SetConfig) (*Set, error) {
	var err error

	todo, err := NewTodo(TodoConfig{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var collectorSet *collector.Set
	{
		c := collector.SetConfig{
			Collectors: []collector.Interface{
				todo,
			},
			Logger: config.Logger,
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
