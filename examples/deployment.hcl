variable "namespace" {
  type        = "string"
  default     = "default"
  description = "Kubernetes namespace"
}

variable "replicas" {
  type    = "number"
  default = 3
  description = "Number of replicas"
}

variable "image_tag" {
  type    = "string"
  default = "latest"
  description = "Docker image tag"
}

resource "kubernetes_deployment" "web" {
  replicas = 3
  
  selector = {
    matchLabels = {
      app = "web"
    }
  }
  
  template = {
    metadata = {
      labels = {
        app = "web"
      }
    }
    spec = {
      containers = [
        {
          name  = "web"
          image = "nginx:latest"
          ports = [
            {
              containerPort = 80
            }
          ]
        }
      ]
    }
  }
}

resource "kubernetes_service" "web" {
  type = "ClusterIP"
  
  selector = {
    app = "web"
  }
  
  ports = [
    {
      port       = 80
      targetPort = 80
      protocol   = "TCP"
    }
  ]
}
