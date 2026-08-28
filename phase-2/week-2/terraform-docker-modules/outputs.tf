output "container_id" {
  description = "ID of the created Docker container"
  value       = module.web_server.container_id
}

output "container_name" {
  description = "Name of the created Docker container"
  value       = module.web_server.container_name
}

output "application_url" {
  description = "URL for accessing nginx"
  value       = "http://localhost:${var.external_port}"
}
