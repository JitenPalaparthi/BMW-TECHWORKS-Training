resource "docker_image" "app_image" {
  name         = var.image_name
  keep_locally = false
}

resource "docker_container" "app_container" {
  name  = var.container_name
  image = docker_image.app_image.image_id

  ports {
    internal = 80
    external = var.external_port
  }
}
