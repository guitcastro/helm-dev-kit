variable "namespace" {
  type        = "string"
  default     = "microservices"
  description = "Kubernetes namespace for microservices"
}

variable "frontend_replicas" {
  type        = "number"
  default     = 2
  description = "Number of frontend replicas"
}

variable "frontend_image" {
  type        = "string"
  default     = "frontend:latest"
  description = "Frontend container image"
}

resource "kubernetes_deployment" "frontend" {
  replicas = 2
  
  selector = {
    matchLabels = {
      app = "frontend"
    }
  }
  
  template = {
    metadata = {
      labels = {
        app = "frontend"
      }
    }
    spec = {
      containers = [
        {
          name  = "frontend"
          image = "frontend:latest"
          ports = [
            {
              containerPort = 3000
            }
          ]
        }
      ]
    }
  }
}

resource "kubernetes_service" "frontend" {
  type = "ClusterIP"
  
  selector = {
    app = "frontend"
  }
  
  ports = [
    {
      port       = 80
      targetPort = 3000
      protocol   = "TCP"
    }
  ]
}