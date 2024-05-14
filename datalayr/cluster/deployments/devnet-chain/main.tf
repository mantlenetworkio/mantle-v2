provider "aws" {
  region = local.aws_region
}

locals {
  cluster_name = basename(path.cwd)
  aws_region   = "us-east-2"
  vpc_cidr = "10.0.0.0/16"
  azs      = slice(data.aws_availability_zones.available.names, 0, 3)
}

data "aws_availability_zones" "available" {}

module "geth_eks" {
  source = "../modules/eks/"

  cluster_name   = local.cluster_name
  cluster_region = local.aws_region
  cluster_version = "1.24"

  vpc_id = module.vpc.vpc_id
  private_subnets = module.vpc.private_subnets

  managed_node_groups = {
    mg_m5_one = {
      # Node Group configuration
      node_group_name         = "managed-ondemand"
      instance_types          = ["m5.xlarge"]
      disk_size               = 32

      # Node Group scaling configuration
      desired_size = 4
      max_size     = 4
      min_size     = 4

      # Node Group network configuration
      subnet_ids = module.vpc.public_subnets
    }
  }
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "3.14.2"

  name = "${local.cluster_name}-vpc"

  cidr = local.vpc_cidr
  azs  = local.azs

  public_subnets  = [for k, v in local.azs : cidrsubnet(local.vpc_cidr, 4, k)]
  private_subnets = [for k, v in local.azs : cidrsubnet(local.vpc_cidr, 4, k + 10)]

  enable_nat_gateway   = true
  single_nat_gateway   = true
  enable_dns_hostnames = true

  public_subnet_tags = {
    "kubernetes.io/cluster/${local.cluster_name}" = "shared"
    "kubernetes.io/role/elb"                    = 1
  }
}