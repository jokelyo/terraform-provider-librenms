# create a service that checks HTTPS cert expiration
resource "librenms_service" "my_device_cert_expiration" {
  device_id = librenms_device.my_device.id
  name      = "My Service Cert Expiration"
  type      = "http"

  ignore     = false
  parameters = "-C 30,14"
  target     = "myservice.mydomain.com"

}
