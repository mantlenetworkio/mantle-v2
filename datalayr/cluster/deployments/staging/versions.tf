terraform {

  cloud {
    organization = "Layr-Labs"

    workspaces {
      name = "staging"
    }
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.35.0"
    }

    random = {
      source  = "hashicorp/random"
      version = "3.4.3"
    }

    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.14"
    }

    helm = {
      source  = "hashicorp/helm"
      version = ">= 2.4.1"
    }
  }

  required_version = "~> 1.3.3"
}
