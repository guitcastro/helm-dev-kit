variable "app_config" {
  type        = "string"
  default     = "production"
  description = "Application configuration environment"
}

resource "kubernetes_config_map" "app_config" {
  data = {
    environment     = "production"
    log_level      = "info"
    cache_timeout  = "300"
    max_connections = "100"
  }
}

resource "kubernetes_config_map" "nginx_config" {
  data = {
    "nginx.conf" = "server { listen 80; location / { proxy_pass http://frontend:80; } }"
  }
}