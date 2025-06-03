# Manage example device.
resource "librenms_device" "my_server" {
  hostname = "my-server.mydomain.com"

  snmp_v2c = {
    # in a production environment, use something such as vault to manage and provide the secrets value
    community = "public"
  }
}

# Force add an ICMP-only device.
resource "librenms_device" "icmp_only_device" {
  hostname  = "icmp-only-device.mydomain.com"
  force_add = true

  icmp_only = {
    hardware = "Generic"
    os       = "Linux"
  }
}
