package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccK0sctlConfigResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccK0sctlConfigResourceConfig_minimal(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("k0sctl_config.test", "spec.host.0.hooks.0.apply.0.before.0", "ls -la"),
					resource.TestCheckResourceAttr("k0sctl_config.test", "spec.k0s.version", "0.13"),
				),
			},
		},
	})
}

func testAccK0sctlConfigResourceConfig_minimal() string {
	return `
resource "k0sctl_config" "test" {
    metadata {
        name = "test"
    }
    spec {
        k0s {
            version = "0.13"
        }

        host {
            role = "controller"
            ssh {
                address  = "controller1.example.org"
                key_path = "./key.pem"
                user     = "ubuntu"
            }

            hooks {
                apply {
                    before = [ "ls -la", "pwd" ]
                }
            }
        }

        host {
            role = "worker"
            ssh {
                address  = "worker1.example.org"
                key_path = "./key.pem"
                user     = "ubuntu"
            }
        }

        host {
            role = "worker"
            winrm {
                address  = "windowsworker1.example.org"
                user     = "ubuntu"
                password = "my-win-password"
            }
        }

    }
}
`
}
