variable "backend_replicas" {
  type        = "number"
  default     = 3
  description = "Number of backend replicas"
}

variable "backend_image" {
  type        = "string"
  default     = "backend:latest"
  description = "Backend container image"
}

variable "database_url" {
  type        = "string"
  default     = "postgresql://db:5432/myapp"
  description = "Database connection URL"
}

resource "kubernetes_deployment" "backend" {
  replicas = 3
  
  selector = {
    matchLabels = {
      app = "backend"
    }
  }
  
  template = {
    metadata = {
      labels = {
        app = "backend"
      }
    }
    spec = {
      containers = [
        {
          name  = "backend"
          image = "backend:latest"
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

resource "kubernetes_service" "backend" {
  type = "ClusterIP"
  
  selector = {
    app = "backend"
  }
  
  ports = [
    {
      port       = 8080
      targetPort = 8080
      protocol   = "TCP"
    }
  ]
}