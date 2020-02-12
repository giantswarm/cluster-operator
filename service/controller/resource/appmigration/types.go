package appmigration

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
)

type Patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

type response struct {
	ChartConfigs []v1alpha1.ChartConfig
	Error        error
}
