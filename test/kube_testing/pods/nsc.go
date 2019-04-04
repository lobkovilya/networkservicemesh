package pods

import (
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
)

func NSCPodWebhook(name string, node *v1.Node) *v1.Pod {
	pod := &v1.Pod{
		ObjectMeta: v12.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				"ns.networkservicemesh.io": "icmp-responder?app=icmp",
			},
		},
		TypeMeta: v12.TypeMeta{
			Kind: "Deployment",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            "alpine-img",
					Image:           "alpine:latest",
					ImagePullPolicy: v1.PullIfNotPresent,
					Command: []string{
						"tail", "-f", "/dev/null",
					},
				},
			},
		},
	}
	if node != nil {
		pod.Spec.NodeSelector = map[string]string{
			"kubernetes.io/hostname": node.Labels["kubernetes.io/hostname"],
		}
	}
	return pod
}

func nscPod(name string, node *v1.Node, env map[string]string, image string) *v1.Pod {
	ht := new(v1.HostPathType)
	*ht = v1.HostPathDirectoryOrCreate

	nsc_container := containerMod(&v1.Container{
		Name:            "nsc",
		Image:           image,
		ImagePullPolicy: v1.PullIfNotPresent,
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{
				"networkservicemesh.io/socket": resource.NewQuantity(1, resource.DecimalSI).DeepCopy(),
			},
			Requests: nil,
		},
	})
	for k, v := range env {
		logrus.Infof("Setting %v: %v", k, v)
		nsc_container.Env = append(nsc_container.Env,
			v1.EnvVar{
				Name:  k,
				Value: v,
			})
	}

	pod := &v1.Pod{
		ObjectMeta: v12.ObjectMeta{
			Name: name,
		},
		TypeMeta: v12.TypeMeta{
			Kind: "Deployment",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            "alpine-img",
					Image:           "alpine:latest",
					ImagePullPolicy: v1.PullIfNotPresent,
					Command: []string{
						"tail", "-f", "/dev/null",
					},
				},
			},
			InitContainers: []v1.Container{
				nsc_container,
			},
		},
	}
	if node != nil {
		pod.Spec.NodeSelector = map[string]string{
			"kubernetes.io/hostname": node.Labels["kubernetes.io/hostname"],
		}
	}
	return pod
}

func NSCPod(name string, node *v1.Node, env map[string]string) *v1.Pod {
	return nscPod(name, node, env, "networkservicemesh/nsc:latest")
}

func GreedyNSCPod(name string, node *v1.Node, env map[string]string, numOfConnections int) *v1.Pod {
	env["NUM_OF_CONNECTIONS"] = strconv.Itoa(numOfConnections)
	return nscPod(name, node, env, "networkservicemesh/greedy-nsc:latest")
}
