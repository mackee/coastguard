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
    *   Set the values for the variables defined in `variables.tf`:
        *   `region`: AWS region.
        *   `project_name`: Project name used for resource names and tags (default: `coastguard-demo`).
        *   `repo`: Repository name used for tags (default: `github.com/mackee/coastguard`).
        *   `allowed_domains`: (Optional) List of Google Workspace domains (`hd` claim) allowed access (e.g., `["example.com"]`).
        *   `allowed_emails`: (Optional) List of email addresses allowed access (e.g., `["guest@gmail.com"]`). Only verified emails are matched.
        *   Users matching either `allowed_domains` or `allowed_emails` are allowed. If both are empty, all authenticated users are allowed. To allow only specific users, set `allowed_emails` alone.
    *   The `task plan` / `task apply` commands pass `region`, `project_name` and `repo` from the `AWS_REGION`, `PROJECT_NAME` and `REPO` environment variables. Set other variables via a `.tfvars` file or environment variables (`TF_VAR_variable_name`, e.g., `TF_VAR_allowed_emails='["guest@gmail.com"]'`).
    *   Place the following files in this `terraform/` directory. Their contents are stored in SSM Parameter Store. **Do not commit them to version control** (they are listed in `.gitignore`).
        *   `oidc.json`: The OAuth client JSON downloaded from your OIDC provider (Google). The client ID and secret are read from it.
        *   `private_key.pem` / `public_key.pem`: The key pair for CloudFront signed cookies (`task generate-public-key` generates them).
    *   The session secret is generated automatically.

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
