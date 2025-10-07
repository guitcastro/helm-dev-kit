variable "app_name" {
  type        = "string"
  default     = "myapp"
  description = "Application name"
}

variable "config_data" {
  type    = "string"
  default = "production"
  description = "Configuration data"
}

resource "kubernetes_config_map" "app_config" {
  data = {
    environment = "production"
    log_level   = "info"
  }
}

resource "kubernetes_deployment" "app" {
  replicas = 2
  
  selector = {
    matchLabels = {
      app = "myapp"
    }
  }
  
  template = {
    metadata = {
      labels = {
        app = "myapp"
      }
    }
    spec = {
      containers = [
        {
          name  = "app"
          image = "myapp:v1.0.0"
          env = [
            {
              name = "LOG_LEVEL"
              valueFrom = {
                configMapKeyRef = {
                  name = "app-config"
                  key  = "log_level"
                }
              }
            }
          ]
          ports = [
            {
              containerPort = 8080
            }
          ]
        }
      ]
    }
  }
}
