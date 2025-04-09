resource "aws_lambda_function" "coastguard" {
  function_name = "coastguard"
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  filename      = data.archive_file.coastguard.output_path
  role          = aws_iam_role.coastguard.arn
  architectures = ["arm64"]
  memory_size   = 128
  timeout       = 3
  environment {
    variables = {
      SSMWRAP_PATHS          = "/coastguard/*"
      OIDC_ISSUER            = "https://accounts.google.com"
      PRESIGN_COOKIE_AGE     = "10h"
      RESTRICT_PATH          = "/"
      ALLOWED_DOMAINS        = ""
      CLOUDFRONT_KEY_PAIR_ID = aws_cloudfront_public_key.coastguard.id
    }
  }

  logging_config {
    log_group  = aws_cloudwatch_log_group.coastguard.name
    log_format = "Text"
  }
}

resource "aws_lambda_alias" "coastguard_current" {
  name             = "current"
  function_name    = aws_lambda_function.coastguard.arn
  function_version = "$LATEST"

  lifecycle {
    ignore_changes = all
  }
}

resource "aws_lambda_function_url" "coastguard" {
  function_name      = aws_lambda_function.coastguard.function_name
  authorization_type = "AWS_IAM"
  qualifier          = aws_lambda_alias.coastguard_current.name
}

data "archive_file" "coastguard" {
  type        = "zip"
  source_file = "${path.module}/bootstrap"
  output_path = "${path.module}/coastguard.zip"

}

resource "aws_lambda_permission" "allow_cloudfront_coastguard" {
  statement_id  = "AllowCloudFrontServicePrincipalCoastguard"
  action        = "lambda:InvokeFunctionUrl"
  function_name = aws_lambda_function.coastguard.function_name
  principal     = "cloudfront.amazonaws.com"
  source_arn    = aws_cloudfront_distribution.main.arn
  qualifier     = aws_lambda_alias.coastguard_current.name
}

resource "aws_iam_role" "coastguard" {
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
  path               = "/service-role/"
}

resource "aws_iam_role_policy_attachment" "coastguard" {
  policy_arn = aws_iam_policy.coastguard.arn
  role       = aws_iam_role.coastguard.name
}

resource "aws_iam_policy" "coastguard" {
  policy = data.aws_iam_policy_document.coastguard.json
  path   = "/service-role/"
}

data "aws_iam_policy_document" "coastguard" {
  statement {
    actions   = ["logs:CreateLogStream", "logs:PutLogEvents"]
    effect    = "Allow"
    resources = ["${aws_cloudwatch_log_group.coastguard.arn}:*"]
  }
  statement {
    actions   = ["ssm:GetParametersByPath"]
    effect    = "Allow"
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_cloudwatch_log_group" "coastguard" {
  name              = "/aws/lambda/coastguard"
  retention_in_days = 30
}

resource "aws_cloudfront_public_key" "coastguard" {
  name        = "coastguard"
  encoded_key = file("public_key.pem")
}

resource "aws_cloudfront_key_group" "coastguard" {
  name  = "coastguard"
  items = [aws_cloudfront_public_key.coastguard.id]
}

resource "aws_cloudfront_origin_access_control" "coastguard" {
  name                              = "coastguard"
  origin_access_control_origin_type = "lambda"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

locals {
  oidc_file_content  = file("${path.module}/oidc.json")
  oidc_file_exists   = fileexists("${path.module}/oidc.json")
  oidc_data          = local.oidc_file_exists ? jsondecode(local.oidc_file_content) : {}
  oidc_client_id     = local.oidc_file_exists ? try(local.oidc_data.web.client_id, local.oidc_data.client_id) : null
  oidc_client_secret = local.oidc_file_exists ? try(local.oidc_data.web.client_secret, local.oidc_data.client_secret) : null
}

resource "aws_ssm_parameter" "coastguard_CLIENT_ID" {
  name             = "/coastguard/CLIENT_ID"
  type             = "String"
  value_wo         = local.oidc_client_id
  value_wo_version = 1
}

resource "aws_ssm_parameter" "coastguard_CLIENT_SECRET" {
  name             = "/coastguard/CLIENT_SECRET"
  type             = "SecureString"
  value_wo         = local.oidc_client_secret
  value_wo_version = 1
}

resource "aws_ssm_parameter" "coastguard_SESSION_SECRET" {
  name             = "/coastguard/SESSION_SECRET"
  type             = "SecureString"
  value_wo         = data.external.random_bytes.result["rand"]
  value_wo_version = 1
}

data "external" "random_bytes" {
  program = ["bash", "-c", "openssl rand -base64 32 | jq -R '{rand: (. | sub(\"=+$\"; \"\"))}'"]
}

resource "aws_ssm_parameter" "coastguard_BASE_URL" {
  name  = "/coastguard/BASE_URL"
  type  = "String"
  value = "https://${aws_cloudfront_distribution.main.domain_name}/"
}

resource "aws_ssm_parameter" "coastguard_SIGN_PRIVATE_KEY" {
  name             = "/coastguard/SIGN_PRIVATE_KEY"
  type             = "SecureString"
  value_wo         = fileexists("private_key.pem") ? sensitive(file("private_key.pem")) : null
  value_wo_version = 1
}
