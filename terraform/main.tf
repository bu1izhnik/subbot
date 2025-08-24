terraform {
  required_version = ">= 1.5.0"
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.28"
    }
    kubectl = {
      source  = "gavinbunney/kubectl"
      version = "~> 1.14"
    }
  }
}

resource "null_resource" "kind_cluster" {
  triggers = {
    cluster_name = var.cluster_name
  }

  provisioner "local-exec" {
    command = "kind create cluster --name ${self.triggers.cluster_name} --wait 120s"
  }

  provisioner "local-exec" {
    when    = destroy
    command = "kind delete cluster --name ${self.triggers.cluster_name}"
  }
}

provider "kubernetes" {
  config_path    = "~/.kube/config"
  config_context = "kind-${var.cluster_name}"
}

provider "kubectl" {
  config_path      = "~/.kube/config"
  config_context   = "kind-${var.cluster_name}"
  load_config_file = true
}

resource "kubernetes_namespace" "db" {
  metadata {
    name = "db"
  }
  depends_on = [null_resource.kind_cluster]
}
resource "kubernetes_namespace" "bot" {
  metadata {
    name = "bot"
  }
  depends_on = [null_resource.kind_cluster]
}
resource "kubernetes_namespace" "fetcher" {
  metadata {
    name = "fetcher"
  }
  depends_on = [null_resource.kind_cluster]
}
