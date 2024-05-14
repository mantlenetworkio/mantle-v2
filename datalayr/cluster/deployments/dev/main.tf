provider "aws" {
  region = local.aws_region
    assume_role {
    role_arn    = "arn:aws:iam::225727539677:role/terraform-iac"
    external_id = "terraform-iac"
  }
}

locals {
  cluster_name = basename(path.cwd)
  aws_region = "us-east-2"
}

module "eigenda_eks" {
  source = "../modules/eks/"

  cluster_name = local.cluster_name
  cluster_region = local.aws_region
  cluster_version = "1.24"

  instance_type = "m5.xlarge"
  min_size = 4
  max_size = 4
  desired_size = 4
  disk_size = 32
}
