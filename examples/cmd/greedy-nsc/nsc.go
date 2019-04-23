package main

import (
	"fmt"
	"github.com/networkservicemesh/networkservicemesh/pkg/tools"
	"github.com/networkservicemesh/networkservicemesh/sdk/client"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
)

/*
Greedy NSC allows you to create N requests to NSM, where N is specified in NUM_OF_CONNECTIONS
*/

const (
	numOfConnectionsDefault = 10
	numOfConnectionsEnv     = "NUM_OF_CONNECTIONS"
)

func CloseWithLogs(c func() error) {
	if err := c(); err != nil {
		logrus.Error(err)
	}
}

func main() {
	tracer, closer := tools.InitJaeger("nsc")
	opentracing.SetGlobalTracer(tracer)
	defer CloseWithLogs(closer.Close)

	nsmClient, err := client.NewNSMClient(nil, nil)
	if err != nil {
		logrus.Fatalf("Unable to create the NSM client %v", err)
	}

	numOfConnections := numOfConnectionsDefault
	if numOfRequestsEnvValue, ok := os.LookupEnv(numOfConnectionsEnv); ok {
		logrus.Infof("%v = %v", numOfConnectionsEnv, numOfRequestsEnvValue)
		numOfConnections, err = strconv.Atoi(numOfRequestsEnvValue)
		if err != nil {
			logrus.Fatalf("Unable to parse %v: %v", numOfConnectionsEnv, err)
		}
	}

	for i := 0; i < numOfConnections; i++ {
		t1 := time.Now()
		//var conn *connection.Connection
		if _, err := nsmClient.Connect(fmt.Sprintf("nsm-%v", i), "kernel", "Primary interface"); err != nil {
			logrus.Fatalf("Client connect failed with error: %v", err)
		}
		logrus.Infof("Connection nsm-%v established: %v", i, time.Since(t1))

		//logrus.Info("Closing...")
		//if err := nsmClient.Close(conn); err != nil {
		//	logrus.Errorf("Error during connection closing: %v", err)
		//}
		//logrus.Info("Sleeping...")
		//time.Sleep(5 * time.Second)
	}

	logrus.Info("nsm client: initialization is completed successfully")
}
