from jinja2 import Environment, FileSystemLoader

env = Environment(loader=FileSystemLoader("."))

template = env.get_template("config.j2")

data = {
    "port": 443,
    "server_name": "example.com",
    "ssl_enabled": True,
    "ssl_certificate": "/etc/ssl/server.crt",
    "ssl_certificate_key": "/etc/ssl/server.key",
    "backend_host": "10.0.1.20",
    "backend_port": 8080,
    "name":"Jiten"
}

output = template.render(data)

print(output)