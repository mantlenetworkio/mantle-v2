output "cluster_id" {
  description = "The ID of the EKS Cluster"
  value       = module.eks_blueprints.eks_cluster_id
}

output "cluster_endpoint" {
  description = "Endpoint for your Kubernetes API server"
  value       = module.eks_blueprints.eks_cluster_endpoint
}

output "cluster_certificate_authority_data" {
  description = "Base64 encoded certificate data required to communicate with the cluster"
  value       = module.eks_blueprints.eks_cluster_certificate_authority_data
}
