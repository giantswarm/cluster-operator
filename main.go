package main

import (
	"fmt"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/microkit/command"
	microserver "github.com/giantswarm/microkit/server"
	"github.com/giantswarm/microkit/transaction"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"
	"github.com/giantswarm/microstorage/memory"
	"github.com/spf13/viper"

	"github.com/giantswarm/cluster-operator/flag"
	"github.com/giantswarm/cluster-operator/server"
	"github.com/giantswarm/cluster-operator/service"
)

const (
	notAvailable = "n/a"
)

var (
	description            = "The cluster-operator is an in-cluster agent that handles Giant Swarm guest cluster specific resources."
	f           *flag.Flag = flag.New()
	gitCommit              = notAvailable
	name                   = "cluster-operator"
	source                 = "https://github.com/giantswarm/cluster-operator"
)

func main() {
	err := mainWithError()
	if err != nil {
		panic(fmt.Sprintf("%#v\n", err))
	}
}

func mainWithError() (err error) {

	// Create a new logger that is used by all packages.
	var newLogger micrologger.Logger
	{
		loggerConfig := micrologger.DefaultConfig()
		loggerConfig.IOWriter = os.Stdout
		newLogger, err = micrologger.New(loggerConfig)
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
			serviceConfig := service.DefaultConfig()

			serviceConfig.Flag = f
			serviceConfig.Logger = newLogger
			serviceConfig.Viper = v

			serviceConfig.Description = description
			serviceConfig.GitCommit = gitCommit
			serviceConfig.Name = name
			serviceConfig.Source = source

			newService, err = service.New(serviceConfig)
			if err != nil {
				panic(fmt.Sprintf("%#v\n", microerror.Maskf(err, "service.New")))
			}

			go newService.Boot()
		}

		var storage microstorage.Storage
		{
			storage, err = memory.New(memory.DefaultConfig())
			if err != nil {
				panic(fmt.Sprintf("%#v\n", microerror.Maskf(err, "memory.New")))
			}
		}

		var transactionResponder transaction.Responder
		{
			c := transaction.DefaultResponderConfig()
			c.Logger = newLogger
			c.Storage = storage

			transactionResponder, err = transaction.NewResponder(c)
			if err != nil {
				panic(fmt.Sprintf("%#v\n", microerror.Maskf(err, "transaction.NewResponder")))
			}
		}

		// New custom server that bundles microkit endpoints.
		var newServer microserver.Server
		{
			serverConfig := server.DefaultConfig()

			serverConfig.MicroServerConfig.Logger = newLogger
			serverConfig.MicroServerConfig.ServiceName = name
			serverConfig.MicroServerConfig.TransactionResponder = transactionResponder
			serverConfig.MicroServerConfig.Viper = v
			serverConfig.Service = newService

			newServer, err = server.New(serverConfig)
			if err != nil {
				panic(fmt.Sprintf("%#v\n", microerror.Maskf(err, "server.New")))
			}
		}

		return newServer
	}

	// Create a new microkit command that manages operator daemon.
	var newCommand command.Command
	{
		commandConfig := command.DefaultConfig()

		commandConfig.Logger = newLogger
		commandConfig.ServerFactory = newServerFactory

		commandConfig.Description = description
		commandConfig.GitCommit = gitCommit
		commandConfig.Name = name
		commandConfig.Source = source

		newCommand, err = command.New(commandConfig)
		if err != nil {
			return microerror.Maskf(err, "command.New")
		}
	}

	daemonCommand := newCommand.DaemonCommand().CobraCommand()

	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.Address, "", "Address used to connect to Kubernetes. When empty in-cluster config is created.")
	daemonCommand.PersistentFlags().Bool(f.Service.Kubernetes.InCluster, true, "Whether to use the in-cluster config to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CAFile, "", "Certificate authority file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CrtFile, "", "Certificate file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.KeyFile, "", "Key file path to use to authenticate with Kubernetes.")

	newCommand.CobraCommand().Execute()

	return nil
}
