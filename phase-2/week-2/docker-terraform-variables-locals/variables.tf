variable "container_name" {
  description = "Docker container name"
  type        = string
  default     = "nginx"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "external_port" {
  description = "Port exposed on host"
  type        = number
  default     = 8080
}