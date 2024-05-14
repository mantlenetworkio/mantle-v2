output "cluster_id" {
  description = "EKS cluster ID"
  value       = module.eigenda_eks.cluster_id
}

output "cluster_endpoint" {
  description = "Endpoint for EKS control plane"
  value       = module.eigenda_eks.cluster_endpoint
}

output "cluster_certificate_authority_data" {
  description = "Base64 encoded certificate data required to communicate with the cluster"
  value       = module.eigenda_eks.cluster_certificate_authority_data
}

output "region" {
  description = "AWS region"
  value       = local.aws_region
}

