terraform {
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 3.0"
    }
  }
}

provider "docker" {}

module "web_server" {
  source = "./modules/docker-container"

  container_name = var.container_name
  image_name     = var.image_name
  external_port  = var.external_port
}
