# Coastguard

[日本語版 (Japanese Version)](README.ja.md)

Coastguard is an authentication layer placed in front of AWS CloudFront distributions. It uses an OIDC (OpenID Connect) provider (e.g., Google) and CloudFront signed cookies to ensure that only authenticated users can access the origin resources protected by CloudFront. This includes static web assets hosted on S3, as well as other origins like Lambda Function URLs or Application Load Balancers (ALBs) configured with Origin Access Control (OAC).

## Features

*   User authentication via OIDC provider.
*   Access control using CloudFront signed cookies.
*   Optional restriction of access to users with specific email domains.
*   Serverless architecture running on AWS Lambda.
*   Infrastructure management using Terraform.

## How it Works

1.  When an unauthenticated user attempts to access a protected resource, CloudFront redirects them to the Coastguard authentication endpoint (`/__auth/redirect`).
2.  Coastguard initiates the authentication flow with the configured OIDC provider.
3.  The user authenticates with the OIDC provider.
4.  Upon successful authentication (and optional domain validation), Coastguard generates CloudFront signed cookies.
5.  Coastguard redirects the user back to the originally requested resource.
6.  The user's browser presents the signed cookies to CloudFront, which grants access to the protected S3 assets.
7.  A logout endpoint (`/__auth/logout`) is available to clear the cookies.

## Directory Structure

*   `lambda/`: Contains the Go source code deployed as an AWS Lambda function. Handles authentication logic, OIDC integration, signed cookie generation, etc.
*   `terraform/`: Contains the Terraform code for defining and managing the Coastguard AWS infrastructure (Lambda, CloudFront, S3, IAM, SSM, etc.).
*   `tmp/`: Contains context information about the project (used for generating this README).

## Usage

To use Coastguard, you primarily use the code in the `terraform/` directory to deploy the AWS infrastructure.

The Lambda function binary itself is typically expected to be downloaded from the repository's [GitHub Releases](https://github.com/mackee/coastguard/releases) page. You will need to configure Terraform to use the downloaded binary or follow steps to build it from the source.

For detailed deployment instructions, please refer to `terraform/README.md`.
