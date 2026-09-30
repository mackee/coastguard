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
    *   `variables.tf`で定義されている変数の値を設定します:
        *   `region`: AWSリージョン
        *   `project_name`: リソース名やタグに使用されるプロジェクト名 (デフォルト: `coastguard-demo`)
        *   `repo`: タグに使用されるリポジトリ名 (デフォルト: `github.com/mackee/coastguard`)
        *   `allowed_domains`: (オプション) アクセスを許可するGoogle Workspaceドメイン (`hd`クレーム) のリスト (例: `["example.com"]`)
        *   `allowed_emails`: (オプション) アクセスを許可するメールアドレスのリスト (例: `["guest@gmail.com"]`)。検証済みのメールアドレスのみ一致とみなされます
        *   `allowed_domains` と `allowed_emails` のいずれかに一致すれば許可されます。両方が空の場合は認証済みの全ユーザーが許可されます。特定のユーザーのみに絞る場合は `allowed_emails` のみを設定してください。
    *   `task plan` / `task apply` は `region`、`project_name`、`repo` を環境変数 `AWS_REGION`、`PROJECT_NAME`、`REPO` から渡します。その他の変数は `.tfvars` ファイルか環境変数 (`TF_VAR_variable_name`、例: `TF_VAR_allowed_emails='["guest@gmail.com"]'`) で設定してください。
    *   以下のファイルをこの `terraform/` ディレクトリに配置します。内容はSSM Parameter Storeに保存されます。**バージョン管理に含めないでください**（`.gitignore` に登録済みです）。
        *   `oidc.json`: OIDCプロバイダー (Google) からダウンロードしたOAuthクライアントのJSON。クライアントIDとシークレットはここから読み込まれます
        *   `private_key.pem` / `public_key.pem`: CloudFront署名付きCookie用のキーペア (`task generate-public-key` で生成できます)
    *   セッションシークレットは自動生成されます。

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
