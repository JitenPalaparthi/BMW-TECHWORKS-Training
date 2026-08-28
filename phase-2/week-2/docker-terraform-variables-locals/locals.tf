locals {
  image_name = "nginx:latest"
  full_container_name = "${var.container_name}-${var.environment}"
}