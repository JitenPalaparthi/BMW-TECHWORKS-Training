variable "container_name" {
  description = "Docker container name"
  type        = string
  default     = "my-web-server"
}

variable "image_name" {
  description = "Docker image name"
  type        = string
  default     = "nginx:latest"
}

variable "external_port" {
  description = "Host port mapped to container port 80"
  type        = number
  default     = 8080
}
