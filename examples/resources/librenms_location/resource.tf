resource "librenms_location" "test_location" {
  name = "test location"

  fixed_coordinates = true
  latitude          = -45.0862462
  longitude         = 37.4220648
}
