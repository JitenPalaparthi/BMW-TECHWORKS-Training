output "container_name" {
  description = "Created Docker container name"
  value       = docker_container.nginx.name
}

output "container_id" {
  description = "Docker container ID"
  value       = docker_container.nginx.id
}

output "application_url" {
  description = "URL for accessing nginx"

  value = "http://localhost:${var.external_port}"
}
