// +build benchmark

package nsmd_integration_tests

import (
	"github.com/networkservicemesh/networkservicemesh/test/integration/nsmd_test_utils"
	"github.com/networkservicemesh/networkservicemesh/test/kube_testing"
	"github.com/networkservicemesh/networkservicemesh/test/kube_testing/pods"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"strings"
	"testing"
	"time"
)

func TestPlentyNsc(t *testing.T) {
	RegisterTestingT(t)

	//ifaceMap := nsmd_test_utils.IpAddr(nil, nil, "")
	//for k,v := range ifaceMap {
	//	fmt.Println(k, v)
	//}
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
	nodes_setup := nsmd_test_utils.SetupNodes(k8s, nodesCount, defaultTimeout)

	nsmd_test_utils.DeployICMP(k8s, nodes_setup[1].Node, "icmp-1", defaultTimeout)

	//nscList := []*v1.Pod{}
	//for i := 0; i < 50; i++ {
	//	nsc := nsmd_test_utils.DeployNSC(k8s, nodes_setupku	[0].Node, fmt.Sprintf("nsc-%v", i), defaultTimeout, false)
	//	nscList = append(nscList, nsc)
	//}
	//
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
	numOfConnections := 2000
	nsc := k8s.CreatePod(pods.GreedyNSCPod("greedy-nsc", nodes_setup[0].Node, env, numOfConnections))
	Expect(nsc.Name).To(Equal("greedy-nsc"))

	k8s.WaitLogsContains(nsc, "nsc", "nsm client: initialization is completed successfully", defaultTimeout*10)
	nsmIntfCounter := 0
	for k, v := range nsmd_test_utils.IpAddr(k8s, nsc, nsc.Spec.Containers[0].Name) {
		logrus.Infof("%v: %v", k, v)
		if strings.Contains(k, "nsm") {
			nsmIntfCounter++
		}
	}
	Expect(nsmIntfCounter).To(Equal(numOfConnections))
}
