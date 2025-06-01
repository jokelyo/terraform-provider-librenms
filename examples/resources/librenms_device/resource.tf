# Manage example device.
resource "librenms_device" "my_server" {
  hostname = "my-server.mydomain.com"

  snmp_v2c = {
    # in a production environment, use something such as vault to manage and provide the secrets value
    community = "public"
  }
}
