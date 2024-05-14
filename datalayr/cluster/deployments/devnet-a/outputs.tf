output "cluster_name" {
  description = "Kubernetes Cluster Name"
  value       = local.cluster_name
}

output "region" {
  description = "AWS region"
  value       = local.aws_region
}
