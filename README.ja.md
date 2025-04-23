# Coastguard

Coastguardは、AWS CloudFrontディストリビューションの前に配置される認証レイヤーです。OIDC（OpenID Connect）プロバイダー（例: Google）とCloudFront署名付きCookieを使用して、CloudFrontによって保護されているオリジンリソース（S3でホストされる静的Webアセット、Origin Access Control (OAC) で設定されたLambda Function URLやApplication Load Balancer (ALB) など）へのアクセスを認証されたユーザーのみに許可します。

## 機能

*   OIDCプロバイダーによるユーザー認証
*   CloudFront署名付きCookieを使用したアクセス制御
*   特定のメールドメインを持つユーザーのみにアクセスを制限する機能（オプション）
*   AWS Lambda上で動作するサーバーレスアーキテクチャ
*   Terraformによるインフラストラクチャ管理

## 仕組み

1.  未認証ユーザーが保護されたリソースにアクセスしようとすると、CloudFrontはCoastguardの認証エンドポイント (`/__auth/redirect`) にリダイレクトします。
2.  Coastguardは設定されたOIDCプロバイダーとの認証フローを開始します。
3.  ユーザーがOIDCプロバイダーで認証します。
4.  認証成功後（およびオプションのドメイン検証後）、CoastguardはCloudFront署名付きCookieを生成します。
5.  Coastguardはユーザーを元のリクエストされたリソースにリダイレクトします。
6.  ユーザーのブラウザは署名付きCookieをCloudFrontに提示し、CloudFrontは保護されたS3アセットへのアクセスを許可します。
7.  ログアウトエンドポイント (`/__auth/logout`) でCookieをクリアできます。

## ディレクトリ構成

*   `lambda/`: AWS Lambda関数としてデプロイされるGo言語のソースコードが含まれます。認証ロジック、OIDC連携、署名付きCookie生成などを担当します。
*   `terraform/`: CoastguardのAWSインフラストラクチャ（Lambda, CloudFront, S3, IAM, SSMなど）を定義・管理するためのTerraformコードが含まれます。
*   `tmp/`: プロジェクトに関するコンテキスト情報（このREADMEの作成に使用）。

## 使い方

Coastguardを利用するには、主に`terraform/`ディレクトリ以下のコードを使用してAWSインフラストラクチャをデプロイします。

Lambda関数本体のバイナリは、通常、リポジトリの[GitHub Releases](https://github.com/mackee/coastguard/releases)ページからダウンロードして使用することを想定しています。Terraformの設定で、ダウンロードしたバイナリを指定するか、ソースからビルドする手順を実行する必要があります。

詳細なデプロイ手順については、`terraform/README.md`を参照してください。
