provider "kubernetes" {
  host                   = module.eks_blueprints.eks_cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks_blueprints.eks_cluster_certificate_authority_data)
  token                  = data.aws_eks_cluster_auth.this.token
}

provider "helm" {
  kubernetes {
    host                   = module.eks_blueprints.eks_cluster_endpoint
    cluster_ca_certificate = base64decode(module.eks_blueprints.eks_cluster_certificate_authority_data)
    token                  = data.aws_eks_cluster_auth.this.token
  }
}

data "aws_eks_cluster_auth" "this" {
  name = module.eks_blueprints.eks_cluster_id
}

module "eks_blueprints" {
  source = "github.com/aws-ia/terraform-aws-eks-blueprints?ref=v4.13.0"

  # EKS CLUSTER
  cluster_name    = var.cluster_name
  cluster_version = var.cluster_version

  vpc_id             = var.vpc_id
  private_subnet_ids = var.private_subnets

  managed_node_groups = var.managed_node_groups

  platform_teams = {
    datalayr-admin = {
      users = [
        "arn:aws:iam::844660611549:user/Daniel",
        "arn:aws:iam::844660611549:user/Bowen",
        "arn:aws:iam::844660611549:user/Dominick",
        "arn:aws:iam::844660611549:user/Robert",
        "arn:aws:iam::844660611549:user/Gautham",
        "arn:aws:iam::844660611549:user/mpereira"
      ]
    }
  }
}

module "eks_blueprints_kubernetes_addons" {
  source = "github.com/aws-ia/terraform-aws-eks-blueprints//modules/kubernetes-addons?ref=v4.13.0"

  eks_cluster_id       = module.eks_blueprints.eks_cluster_id
  eks_cluster_endpoint = module.eks_blueprints.eks_cluster_endpoint
  eks_oidc_provider    = module.eks_blueprints.oidc_provider
  eks_cluster_version  = module.eks_blueprints.eks_cluster_version

  # EKS Addons
  enable_amazon_eks_aws_ebs_csi_driver = true
}

