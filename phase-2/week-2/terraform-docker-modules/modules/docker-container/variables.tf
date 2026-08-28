variable "container_name" {
  description = "Name of Docker container"
  type        = string
}

variable "image_name" {
  description = "Docker image"
  type        = string
}

variable "external_port" {
  description = "Host port"
  type        = number
}
