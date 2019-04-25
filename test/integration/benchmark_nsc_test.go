// +build benchmark

package nsmd_integration_tests

import (
	"fmt"
	"github.com/networkservicemesh/networkservicemesh/test/integration/nsmd_test_utils"
	"github.com/networkservicemesh/networkservicemesh/test/kube_testing"
	"github.com/networkservicemesh/networkservicemesh/test/kube_testing/pods"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"testing"
	"time"
)

func TestPlentyNsc(t *testing.T) {
	RegisterTestingT(t)

	if testing.Short() {
		t.Skip("Skip, please run without -short")
		return
	}

	k8s, err := kube_testing.NewK8s()
	defer k8s.Cleanup()

	Expect(err).To(BeNil())

	s1 := time.Now()
	k8s.PrepareDefault()
	logrus.Printf("Cleanup done: %v", time.Since(s1))

	nodesCount := 2
	nodesSetup := nsmd_test_utils.SetupNodes(k8s, nodesCount, defaultTimeout)

	for i := 0; i < 1; i++ {
		nsmd_test_utils.DeployICMP(k8s, nodesSetup[1].Node, fmt.Sprintf("icmp-%d", i), defaultTimeout)
	}
	//
	//nscList := []*v1.Pod{}
	//for i := 0; i < 50; i++ {
	//	nsc := nsmd_test_utils.DeployNSC(k8s, nodesSetup[0].Node, fmt.Sprintf("nsc-%v", i), defaultTimeout)
	//	nscList = append(nscList, nsc)
	//}

	//for _, nsc := range nscList {
	//	cidr := nsmd_test_utils.IpAddr(k8s, nsc, nsc.Spec.Containers[0].Name)["nsm"]
	//	srcIp, ipNet, _ := net.ParseCIDR(cidr)
	//	dstIp, _ := prefix_pool.IncrementIP(srcIp, ipNet)
	//	nsmd_test_utils.CheckNSCConfig(k8s, t, nsc, srcIp.String(), dstIp.String())
	//}

	env := map[string]string{
		"OUTGOING_NSC_LABELS": "app=icmp",
		"OUTGOING_NSC_NAME":   "icmp-responder",
	}
	numOfConnections := 500
	nsc := k8s.CreatePod(pods.GreedyNSCPod("greedy-nsc", nodesSetup[0].Node, env, numOfConnections))
	Expect(nsc.Name).To(Equal("greedy-nsc"))

	k8s.WaitLogsContains(nsc, "nsc", "nsm client: initialization is completed successfully", defaultTimeout*10)

	r, _, err := k8s.Exec(nsc, "", "ip", "a")
	Expect(err).To(BeNil())
	fmt.Printf(r)
	//Expect(nsmIntfCounter).To(Equal(numOfConnections))
}
