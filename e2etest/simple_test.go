package e2e

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"go.universe.tf/e2etest/pkg/frr/container"

	frrk8sv1alpha1 "github.com/metallb/frrk8s/api/v1alpha1"
	"github.com/metallb/frrk8stests/pkg/config"
	"github.com/metallb/frrk8stests/pkg/infra"
	"go.universe.tf/e2etest/pkg/ipfamily"
	frrconfig "go.universe.tf/metallb/e2etest/pkg/frr/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	admissionapi "k8s.io/pod-security-admission/api"
)

var _ = ginkgo.Describe("FRR k8s", func() {
	var cs clientset.Interface
	var f *framework.Framework

	ginkgo.AfterEach(func() {

		/*
			if ginkgo.CurrentSpecReport().Failed() {
				k8s.DumpInfo(Reporter, ginkgo.CurrentSpecReport().LeafNodeText)
			}
		*/
	})

	ginkgo.BeforeEach(func() {
		ginkgo.By("Clearing any previous configuration")

		for _, c := range infra.FRRContainers {
			err := c.UpdateBGPConfigFile(frrconfig.Empty)
			framework.ExpectNoError(err)
		}
		err := updater.Clean()
		framework.ExpectNoError(err)
	})

	f = framework.NewDefaultFramework("bgpfrr")
	f.NamespacePodSecurityEnforceLevel = admissionapi.LevelPrivileged

	ginkgo.BeforeEach(func() {
		cs = f.ClientSet
	})

	ginkgo.Context("Basic tests", func() {
		ginkgo.It("establishes session with external frrs", func() {
			router := frrk8sv1alpha1.Router{
				ASN:       infra.FRRK8sASN,
				Neighbors: config.NeighborsForContainers(infra.FRRContainers),
			}
			config := frrk8sv1alpha1.FRRConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: frrk8sv1alpha1.FRRConfigurationSpec{
					Routers: []frrk8sv1alpha1.Router{router},
				},
			}

			for _, c := range infra.FRRContainers {
				err := container.PairWithNodes(cs, c, ipfamily.IPv4)
				framework.ExpectNoError(err)
			}
			err := updater.Update(config)
			framework.ExpectNoError(err)

			nodes, err := cs.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
			framework.ExpectNoError(err)

			for _, c := range infra.FRRContainers {
				validateFRRPeeredWithNodes(nodes.Items, c, ipfamily.IPv4)
			}
		})
	})
})
