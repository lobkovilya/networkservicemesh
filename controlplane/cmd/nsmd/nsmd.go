package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/networkservicemesh/networkservicemesh/controlplane/pkg/model"
	"github.com/networkservicemesh/networkservicemesh/controlplane/pkg/nsm"
	"github.com/networkservicemesh/networkservicemesh/controlplane/pkg/nsmd"
	"github.com/networkservicemesh/networkservicemesh/pkg/tools"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

func main() {
	start := time.Now()

	// Capture signals to cleanup before exiting
	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	tracer, closer := tools.InitJaeger("nsmd")
	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()

	go nsmd.BeginHealthCheck()

	apiRegistry := nsmd.NewApiRegistry()
	serviceRegistry := nsmd.NewServiceRegistry()

	model := model.NewModel() // This is TCP gRPC server uri to access this NSMD via network.
	defer serviceRegistry.Stop()

	excludedPrefixes, err := nsmd.GetExcludedPrefixes(serviceRegistry)
	if err != nil {
		logrus.Fatalf("Error during getting Excluded Prefixes: %v", err)
	}

	manager := nsm.NewNetworkServiceManager(model, serviceRegistry, excludedPrefixes)

	var server nsmd.NSMServer
	// Start NSMD server first, laod local NSE/client registry and only then start dataplane/wait for it and recover active connections.
	if server, err = nsmd.StartNSMServer(model, manager, serviceRegistry, apiRegistry); err != nil {
		logrus.Fatalf("Error starting nsmd service: %+v", err)
		nsmd.SetNSMServerFailed()
	}
	defer server.Stop()

	// Starting dataplene
	logrus.Info("Starting Dataplane registration server...")
	if err := server.StartDataplaneRegistratorServer(); err != nil {
		logrus.Fatalf("Error starting dataplane service: %+v", err)
		nsmd.SetDPServerFailed()
	}

	// Wait for dataplane to be connecting to us.
	if err := manager.WaitForDataplane(nsmd.DataplaneTimeout); err != nil {
		logrus.Errorf("Error waiting for dataplane")
	}

	// Choose a public API listener
	sock, err := apiRegistry.NewPublicListener()
	if err != nil {
		logrus.Errorf("Failed to start Public API server...")
		nsmd.SetAPIServerFailed()
	}

	if err := nsmd.StartAPIServerAt(server, sock); err != nil {
		logrus.Fatalf("Error starting nsmd api service: %+v", err)
		nsmd.SetAPIServerFailed()
	}

	elapsed := time.Since(start)
	logrus.Debugf("Starting NSMD took: %s", elapsed)

	<-c
}
