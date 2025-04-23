# Terraform for Coastguard

このディレクトリには、CoastguardをAWSにデプロイするためのTerraformコードが含まれています。

## 前提条件

*   Terraform CLIがインストールされていること。
*   AWSアカウントと認証情報が設定されていること（例: 環境変数、AWSプロファイル）。
*   OIDCプロバイダー（例: Google）でクライアントIDとシークレットを取得済みであること。
*   CloudFront署名付きCookie用のキーペアが生成されていること（`openssl genrsa -out private_key.pem 2048` と `openssl rsa -pubout -in private_key.pem -out public_key.pem` で生成できます）。

## デプロイ手順

1.  **Lambda関数の準備:**
    *   Terraform の `plan` および `apply` タスク (`terraform/Taskfile.yaml` 内の `download-lambda` タスク経由) は、[GitHub Releases](https://github.com/mackee/coastguard/releases) から適切な `coastguard.zip` ファイルを自動的にダウンロードします。
    *   あるいは、手動で目的のリリースの `coastguard_linux_arm64.zip` ファイルをダウンロードし、`coastguard.zip` にリネームして、この `terraform/` ディレクトリに配置してから、(`task` を使わずに) 直接 `terraform plan` または `terraform apply` を実行することも可能です。
    *   **注意:** Terraformは `./coastguard.zip` にzipファイルがあることを期待しています (`coastguard.tf` 内の `aws_lambda_function.coastguard` リソースの `filename` 引数を参照)。

2.  **Terraform変数の設定:**
    *   `variables.tf`で定義されている変数の値を設定します。主な変数は以下の通りです:
        *   `region`: AWSリージョン
        *   `prefix`: リソース名に使用されるプレフィックス
        *   `oidc_client_id`: OIDCプロバイダーのクライアントID
        *   `oidc_client_secret`: OIDCプロバイダーのクライアントシークレット
        *   `oidc_issuer_url`: OIDCプロバイダーのIssuer URL
        *   `redirect_url`: OIDCプロバイダーに登録したリダイレクトURL (例: `https://<your-cloudfront-domain>/__auth/callback`)
        *   `session_secret`: セッション管理用のランダムなシークレットキー
        *   `allowed_domains`: (オプション) アクセスを許可するメールドメインのリスト (例: `["example.com"]`)
        *   `cloudfront_public_key_pem`: CloudFront署名付きCookie用の公開鍵 (`public_key.pem`の内容)
        *   `cloudfront_private_key_pem`: CloudFront署名付きCookie用の秘密鍵 (`private_key.pem`の内容)
    *   これらの変数は、`.tfvars`ファイルを使用するか、環境変数 (`TF_VAR_variable_name`) として設定するのが一般的です。**機密情報（シークレット、秘密鍵）はバージョン管理に含めないでください。** SSM Parameter StoreやSecrets Managerの使用を検討してください（現在のコードではSSM Parameter Storeに保存するようになっています）。

3.  **Terraformの実行:**
    ```bash
    terraform init
    terraform plan # 実行計画を確認
    terraform apply # インフラストラクチャをデプロイ
    ```

4.  **デプロイ後の設定:**
    *   `terraform apply`の出力からCloudFrontディストリビューションのドメイン名を取得します。
    *   OIDCプロバイダーの設定で、リダイレクトURI (Callback URL) として `https://<CloudFrontドメイン名>/__auth/callback` を登録します。

## クリーンアップ

デプロイしたリソースを削除するには、以下のコマンドを実行します。

```bash
terraform destroy
```

**注意:** S3バケットなど、一部のリソースは手動での削除が必要になる場合があります（特にバケットが空でない場合）。
