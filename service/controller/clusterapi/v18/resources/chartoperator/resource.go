package chartoperator

import (
	"reflect"

	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
)

const (
	// Name is the identifier of the resource.
	Name = "chartoperatorv18"
)

const (
	chart         = "chart-operator-chart"
	channel       = "0-9-stable"
	deployment    = "chart-operator"
	release       = "chart-operator"
	namespace     = "giantswarm"
	desiredStatus = "DEPLOYED"
	failedStatus  = "FAILED"
)

// Config represents the configuration used to create a new chartoperator resource.
type Config struct {
	ApprClient apprclient.Interface
	FileSystem afero.Fs
	Logger     micrologger.Logger

	DNSIP          string
	RegistryDomain string
}

// Resource implements the chartoperator resource.
type Resource struct {
	apprClient apprclient.Interface
	fileSystem afero.Fs
	logger     micrologger.Logger

	dnsIP          string
	registryDomain string
}

// New creates a new configured chartoperator resource.
func New(config Config) (*Resource, error) {
	if config.ApprClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ApprClient must not be empty", config)
	}
	if config.FileSystem == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.FileSystem must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.DNSIP == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.DNSIP must not be empty", config)
	}
	if config.RegistryDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RegistryDomain must not be empty", config)
	}

	newResource := &Resource{
		apprClient: config.ApprClient,
		fileSystem: config.FileSystem,
		logger:     config.Logger,

		dnsIP:          config.DNSIP,
		registryDomain: config.RegistryDomain,
	}

	return newResource, nil
}

// Name returns name of the Resource.
func (r *Resource) Name() string {
	return Name
}

func toResourceState(v interface{}) (ResourceState, error) {
	if v == nil {
		return ResourceState{}, nil
	}

	resourceState, ok := v.(*ResourceState)
	if !ok {
		return ResourceState{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", resourceState, v)
	}

	return *resourceState, nil
}

func shouldUpdate(currentState, desiredState ResourceState) bool {
	if currentState.ReleaseVersion != "" && currentState.ReleaseVersion != desiredState.ReleaseVersion {
		// ReleaseVersion has changed for the channel so we need to update the Helm
		// Release.
		return true
	}

	if !reflect.DeepEqual(currentState.ChartValues, desiredState.ChartValues) {
		return true
	}

	if currentState.ReleaseStatus == failedStatus {
		// Release status is failed so do force upgrade to attempt to fix it.
		return true
	}

	return false
}
