# Terraform for Coastguard

[日本語版 (Japanese Version)](README.ja.md)

This directory contains the Terraform code to deploy Coastguard on AWS.

## Prerequisites

*   Terraform CLI installed.
*   AWS account and credentials configured (e.g., via environment variables, AWS profile).
*   Client ID and Secret obtained from your OIDC provider (e.g., Google).
*   A key pair generated for CloudFront signed cookies (you can generate them using `openssl genrsa -out private_key.pem 2048` and `openssl rsa -pubout -in private_key.pem -out public_key.pem`).

## Deployment Steps

1.  **Prepare the Lambda Function:**
    *   The Terraform `plan` and `apply` tasks (via the `download-lambda` task in `terraform/Taskfile.yaml`) will automatically download the correct `coastguard.zip` file from the [GitHub Releases](https://github.com/mackee/coastguard/releases).
    *   Alternatively, you can manually download the `coastguard_linux_arm64.zip` file from the desired release, rename it to `coastguard.zip`, and place it in this `terraform/` directory before running `terraform plan` or `terraform apply` directly (without using `task`).
    *   **Note:** Terraform expects the zip file at `./coastguard.zip` (see the `filename` argument in the `aws_lambda_function.coastguard` resource within `coastguard.tf`).

2.  **Configure Terraform Variables:**
    *   Set the values for the variables defined in `variables.tf`. Key variables include:
        *   `region`: AWS region.
        *   `prefix`: Prefix used for resource names.
        *   `oidc_client_id`: Client ID from your OIDC provider.
        *   `oidc_client_secret`: Client Secret from your OIDC provider.
        *   `oidc_issuer_url`: Issuer URL of your OIDC provider.
        *   `redirect_url`: The redirect URL registered with your OIDC provider (e.g., `https://<your-cloudfront-domain>/__auth/callback`).
        *   `session_secret`: A random secret key for session management.
        *   `allowed_domains`: (Optional) List of email domains allowed access (e.g., `["example.com"]`).
        *   `cloudfront_public_key_pem`: Public key content for CloudFront signed cookies (content of `public_key.pem`).
        *   `cloudfront_private_key_pem`: Private key content for CloudFront signed cookies (content of `private_key.pem`).
    *   It's common practice to set these variables using a `.tfvars` file or environment variables (`TF_VAR_variable_name`). **Do not commit sensitive information (secrets, private keys) to version control.** Consider using SSM Parameter Store or Secrets Manager (the current code is configured to store them in SSM Parameter Store).

3.  **Run Terraform:**
    ```bash
    terraform init
    terraform plan # Review the execution plan
    terraform apply # Deploy the infrastructure
    ```

4.  **Post-Deployment Configuration:**
    *   Obtain the CloudFront distribution domain name from the `terraform apply` output.
    *   Register the redirect URI (Callback URL) `https://<CloudFront-Domain-Name>/__auth/callback` in your OIDC provider's settings.

## Cleanup

To remove the deployed resources, run:

```bash
terraform destroy
```

**Note:** Some resources, like S3 buckets, might require manual deletion, especially if they are not empty.
