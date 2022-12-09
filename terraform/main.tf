terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.45.0"
    }
  }
}

provider "aws" {
  profile = "default"
  region  = "ap-northeast-1"
}

locals {
  app = "go-s3-zipper"
}

data "aws_region" "default" {}

resource "aws_s3_bucket" "image_bucket" {
  bucket_prefix = "${local.app}-images"
  tags          = { App = local.app }
}

resource "aws_s3_bucket_acl" "image_bucket_acl" {
  bucket = aws_s3_bucket.image_bucket.id
  acl    = "public-read"
}

output "image_bucket" {
  value = aws_s3_bucket.image_bucket.bucket
}

resource "aws_s3_bucket" "upload_bucket" {
  bucket_prefix = "${local.app}-uploads"
  tags          = { App = local.app }
}

resource "aws_s3_bucket_acl" "import_bucket_acl" {
  bucket = aws_s3_bucket.upload_bucket.id
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "import_bucket_lifecycle" {
  bucket = aws_s3_bucket.upload_bucket.id

  rule {
    id     = "import-rule"
    status = "Enabled"

    expiration {
      days = 1
    }
  }
}

output "upload_bucket" {
  value = aws_s3_bucket.upload_bucket.bucket
}
