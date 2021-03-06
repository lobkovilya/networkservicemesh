// +build usecase

package nsmd_integration_tests

import (
	"github.com/networkservicemesh/networkservicemesh/test/kube_testing/pods"
	"testing"
	"time"

	"github.com/networkservicemesh/networkservicemesh/test/integration/nsmd_test_utils"
	"github.com/networkservicemesh/networkservicemesh/test/kube_testing"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

func TestNSCAndICMPLocal(t *testing.T) {
	RegisterTestingT(t)

	if testing.Short() {
		t.Skip("Skip, please run without -short")
		return
	}

	testNSCAndICMP(t, 1, false, false)
}

func TestNSCAndICMPRemote(t *testing.T) {
	RegisterTestingT(t)

	if testing.Short() {
		t.Skip("Skip, please run without -short")
		return
	}

	testNSCAndICMP(t, 2, false, false)
}

func TestNSCAndICMPWebhookLocal(t *testing.T) {
	RegisterTestingT(t)

	if testing.Short() {
		t.Skip("Skip, please run without -short")
		return
	}

	testNSCAndICMP(t, 1, true, false)
}

func TestNSCAndICMPWebhookRemote(t *testing.T) {
	RegisterTestingT(t)

	if testing.Short() {
		t.Skip("Skip, please run without -short")
		return
	}

	testNSCAndICMP(t, 2, true, false)
}

func TestNSCAndICMPLocalVeth(t *testing.T) {
	RegisterTestingT(t)

	if testing.Short() {
		t.Skip("Skip, please run without -short")
		return
	}

	testNSCAndICMP(t, 1, false, true)
}

func TestNSCAndICMPRemoteVeth(t *testing.T) {
	RegisterTestingT(t)

	if testing.Short() {
		t.Skip("Skip, please run without -short")
		return
	}

	testNSCAndICMP(t, 2, false, true)
}

/**
If passed 1 both will be on same node, if not on different.
*/
func testNSCAndICMP(t *testing.T, nodesCount int, useWebhook bool, disableVHost bool) {
	k8s, err := kube_testing.NewK8s()
	defer k8s.Cleanup()

	Expect(err).To(BeNil())

	s1 := time.Now()
	k8s.PrepareDefault()
	logrus.Printf("Cleanup done: %v", time.Since(s1))

	config := []*pods.NSMgrPodConfig{}
	for i := 0; i< nodesCount;i++ {
		cfg := &pods.NSMgrPodConfig{}
		if disableVHost {
			cfg.DataplaneVariables = map[string]string{}
			cfg.DataplaneVariables["DATAPLANE_ALLOW_VHOST"] = "false"
		}
		config = append(config, cfg)
	}
	nodes_setup := nsmd_test_utils.SetupNodesConfig(k8s, nodesCount, defaultTimeout, config)

	// Run ICMP on latest node
	_ = nsmd_test_utils.DeployICMP(k8s, nodes_setup[nodesCount-1].Node, "icmp-responder-nse-1", defaultTimeout)

	nscPodNode := nsmd_test_utils.DeployNSC(k8s, nodes_setup[0].Node, "nsc-1", defaultTimeout, useWebhook)

	var nscInfo *nsmd_test_utils.NSCCheckInfo

	failures := InterceptGomegaFailures(func() {
		nscInfo = nsmd_test_utils.CheckNSC(k8s, t, nscPodNode)
	})
	// Do dumping of container state to dig into what is happened.
	if len(failures) > 0 {
		logrus.Errorf("Failures: %v", failures)
		nsmd_test_utils.PrintLogs(k8s, nodes_setup)
		nscInfo.PrintLogs()

		t.Fail()
	}
}
