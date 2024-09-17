package netbox

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/fbreckle/go-netbox/netbox/client"
	"github.com/fbreckle/go-netbox/netbox/client/circuits"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccNetboxCircuitProviderNetworkDependencies(testName string, testSlug string) string {
	return fmt.Sprintf(`
resource "netbox_circuit_provider" "test" {
	name = "%[1]s"
	slug = "%[2]s"
}
`, testName, testSlug)
}
func TestAccNetboxCircuitProviderNetwork_basic(t *testing.T) {
	testSlug := "circuit_prov_network"
	testName := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxCircuitProviderNetworkDependencies(testName, randomSlug) + fmt.Sprintf(`
resource "netbox_circuit_provider_network" "test" {
  name = "%[1]s"
  description = "description test"
  provider_id = netbox_circuit_provider.test.id
  service_id = "service_id test"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_circuit_provider_network.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_circuit_provider_network.test", "description", "description test"),
					resource.TestCheckResourceAttr("netbox_circuit_provider_network.test", "service_id", "service_id test"),
					resource.TestCheckResourceAttrPair("netbox_circuit_provider_network.test", "provider_id", "netbox_circuit_provider.test", "id"),
				),
			},
			{
				Config: testAccNetboxCircuitProviderNetworkDependencies(testName, randomSlug) + fmt.Sprintf(`
resource "netbox_circuit_provider_network" "test" {
  name = "%[1]s"
  status = "active"
  description = "description test"
  provider_id = netbox_circuit_provider.test.id
	service_id = "service_id test"
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_circuit_provider_network.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_circuit_provider_network.test", "description", "description test"),
					resource.TestCheckResourceAttrPair("netbox_circuit_provider_network.test", "provider_id", "netbox_circuit_provider.test", "id"),
					resource.TestCheckResourceAttr("netbox_circuit_provider_network.test", "service_id", "service_id test"),
				),
			},
			{
				ResourceName:      "netbox_circuit.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_circuit_provider_network", &resource.Sweeper{
		Name:         "netbox_circuit_provider_network",
		Dependencies: []string{},
		F: func(region string) error {
			m, err := sharedClientForRegion(region)
			if err != nil {
				return fmt.Errorf("Error getting client: %s", err)
			}
			api := m.(*client.NetBoxAPI)
			params := circuits.NewCircuitsProviderNetworksListParams()
			res, err := api.Circuits.CircuitsProviderNetworksList(params, nil)
			if err != nil {
				return err
			}
			for _, ProviderNetwork := range res.GetPayload().Results {
				if strings.HasPrefix(*ProviderNetwork.Name, testPrefix) {
					deleteParams := circuits.NewCircuitsProviderNetworksDeleteParams().WithID(ProviderNetwork.ID)
					_, err := api.Circuits.CircuitsProviderNetworksDelete(deleteParams, nil)
					if err != nil {
						return err
					}
					log.Print("[DEBUG] Deleted a circuit provider network")
				}
			}
			return nil
		},
	})
}
