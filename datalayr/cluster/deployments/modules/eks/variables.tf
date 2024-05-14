variable "cluster_version" {
  description = "Kubernetes `<major>.<minor>` version to use for the EKS cluster (i.e.: `1.23`)"
  type        = string
  default     = "1.23"
}

variable "cluster_name" {
  description = "The name to use for all the cluster resources"
  type        = string
  default     = ""
}

variable "cluster_timeouts" {
  description = "Create, update, and delete timeout configurations for the cluster"
  type        = map(string)
  default     = {}
}

variable "cluster_region" {
  description = "Region that the cluster is hosted in"
  type        = string
  default     = ""
}

variable "vpc_id" {
  description = "vpc ID"
  type = string
}

variable "private_subnets" {
  description = "Private subnets in the VPC"
  type = any
}

variable "managed_node_groups" {
  description = "Managed node groups for this cluster"
  type = any
}