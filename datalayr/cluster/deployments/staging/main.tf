provider "aws" {
  region = local.aws_region
}

locals {
  cluster_name = "staging"
  aws_region   = "us-west-2"
}

module "staging_eks" {
  source = "../modules/eks/"

  cluster_name   = local.cluster_name
  cluster_region = local.aws_region

  instance_type = "m5.xlarge"
  min_size      = 12
  max_size      = 12
  desired_size  = 12
  disk_size     = 400
}
