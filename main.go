package main

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/microkit/command"
	microserver "github.com/giantswarm/microkit/server"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/viper"

	"github.com/giantswarm/cluster-operator/flag"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/server"
	"github.com/giantswarm/cluster-operator/service"
)

var (
	f = flag.New()
)

func main() {
	err := mainE()
	if err != nil {
		panic(fmt.Sprintf("%#v\n", err))
	}
}

func mainE() error {
	var err error

	ctx := context.Background()

	// Create a new logger that is used by all packages.
	var newLogger micrologger.Logger
	{
		newLogger, err = micrologger.New(micrologger.Config{})
		if err != nil {
			return microerror.Maskf(err, "micrologger.New")
		}
	}

	// Define server factory to create the custom server once all command line
	// flags are parsed and all microservice configuration is processed.
	newServerFactory := func(v *viper.Viper) microserver.Server {
		// New custom service implements the business logic.
		var newService *service.Service
		{
			serviceConfig := service.Config{
				Flag:   f,
				Logger: newLogger,
				Viper:  v,
			}

			newService, err = service.New(serviceConfig)
			if err != nil {
				panic(fmt.Sprintf("%#v\n", microerror.Maskf(err, "service.New")))
			}

			go newService.Boot(ctx)
		}

		// New custom server that bundles microkit endpoints.
		var newServer microserver.Server
		{
			c := server.Config{
				Logger:  newLogger,
				Service: newService,
				Viper:   v,

				ProjectName: project.Name(),
			}

			newServer, err = server.New(c)
			if err != nil {
				panic(fmt.Sprintf("%#v\n", microerror.Maskf(err, "server.New")))
			}
		}

		return newServer
	}

	// Create a new microkit command that manages operator daemon.
	var newCommand command.Command
	{
		c := command.Config{
			Logger:        newLogger,
			ServerFactory: newServerFactory,

			Description: project.Description(),
			GitCommit:   project.GitSHA(),
			Name:        project.Name(),
			Source:      project.Source(),
			Version:     project.Version(),
		}

		newCommand, err = command.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	daemonCommand := newCommand.DaemonCommand().CobraCommand()

	daemonCommand.PersistentFlags().String(f.Guest.Cluster.Calico.CIDR, "", "Prefix length for the CIDR block used by Calico.")
	daemonCommand.PersistentFlags().String(f.Guest.Cluster.Calico.Subnet, "", "Network address for the CIDR block used by Calico.")
	daemonCommand.PersistentFlags().String(f.Guest.Cluster.Kubernetes.API.ClusterIPRange, "", "CIDR Range for Pods in cluster.")
	daemonCommand.PersistentFlags().String(f.Guest.Cluster.Kubernetes.API.Domain, "", "Internal Kubernetes domain.")

	daemonCommand.PersistentFlags().String(f.Guest.Cluster.Vault.Certificate.TTL, "", "Vault certificate TTL.")

	daemonCommand.PersistentFlags().String(f.Service.ClusterService.Address, "http://localhost:8089", "http://<host>:<port> of the cluster service.")

	daemonCommand.PersistentFlags().String(f.Service.Image.Registry.Domain, "quay.io", "Image registry.")

	daemonCommand.PersistentFlags().String(f.Service.KubeConfig.Secret.Namespace, "giantswarm", "The namespace where kubeconfig secrets are located.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.Address, "", "Address used to connect to Kubernetes. When empty in-cluster config is created.")
	daemonCommand.PersistentFlags().Bool(f.Service.Kubernetes.InCluster, true, "Whether to use the in-cluster config to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CAFile, "", "Certificate authority file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CrtFile, "", "Certificate file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.KeyFile, "", "Key file path to use to authenticate with Kubernetes.")

	daemonCommand.PersistentFlags().String(f.Service.Provider.Kind, "", "Provider of the installation. One of aws, azure, kvm.")

	daemonCommand.PersistentFlags().String(f.Service.Release.App.Config.Default, "", "Default properties for app in YAML format.")
	daemonCommand.PersistentFlags().String(f.Service.Release.App.Config.Override, "", "Overriding properties for app in YAML format.")

	newCommand.CobraCommand().Execute()

	return nil
}
