# AGENTS.md

## 目的
- `security-review` は GitHub Actions 向けの Docker Action です。
- PR の変更ファイルに対して Semgrep / Gitleaks を実行し、GitHub PR コメントや RISKEN 連携を行います。

## 技術スタック
- Go `1.25.5`
- Docker Action: [`action.yaml`](./action.yaml)
- Lint / Build Workflow: [`.github/workflows/`](./.github/workflows)

## ディレクトリ構成
- `main.go`
  - エントリーポイント
- `pkg/cmd`
  - CLI 定義
- `pkg/review`
  - GitHub イベント読込、PR コメント投稿、RISKEN 連携の中心ロジック
- `pkg/scanner`
  - Semgrep / Gitleaks のスキャン処理
- `pkg/mocks`
  - mockery 生成物。基本的に手編集しない
- `test/github-event`
  - GitHub event のテストデータ
- `test/review-code`
  - 動作確認用コード

## 変更時の基本方針
- まず `pkg/review` と `pkg/scanner` の責務をまたがせすぎない
- スキャンロジックの変更は、できるだけ `pkg/scanner` に閉じ込める
- GitHub API / RISKEN API まわりの変更は `pkg/review` 側で扱う
- 既存のテーブルドリブンテストの書き方に合わせる
- `for _, tt := range tests` のサブテストでは、必要に応じて `tt := tt` を入れる

## 生成物
- `pkg/mocks/*` は mockery による生成物です
- interface を変えたら `make generate-mock` を使って再生成してください

## よく触るファイル
- [`pkg/review/review.go`](./pkg/review/review.go)
  - 全体の実行フロー
- [`pkg/review/review_gihub.go`](./pkg/review/review_gihub.go)
  - GitHub event / PR コメント処理
- [`pkg/review/review_risken.go`](./pkg/review/review_risken.go)
  - RISKEN 登録処理
- [`pkg/scanner/scan_semgrep.go`](./pkg/scanner/scan_semgrep.go)
  - Semgrep 実行
- [`pkg/scanner/scan_gitleaks.go`](./pkg/scanner/scan_gitleaks.go)
  - Gitleaks 実行

## 開発コマンド
- ツール導入: `make install`
- mock 再生成: `make generate-mock`
- lint: `make lint`
- テスト: `go test ./...`
- Docker build: `make build`
- ローカル実行: `make run`

## Docker / Action の注意
- `action.yaml` は `docker://ssgca/security-review:v1` を参照します
- workflow を変えるだけでは Action 利用先には反映されません
- Action の挙動変更を配布するには Docker image の再 push が必要です

## イメージ公開
- ローカル build は `ssgca/risken-review:$(TAG)` を使います
- 配布用 push は `ssgca/security-review:$(TAG)` を使います
- push は `make push TAG=v1` を利用します
- `make push` は `linux/amd64,linux/arm64` の multi-arch image を push します

## CI
- 現在の GitHub Actions runner は `ubuntu-24.04-arm`
- lint workflow では `golangci-lint` と `deadcode` を使います
- Go バージョンを変えたときは、少なくとも以下を一緒に確認してください
  - `go.mod`
  - `Dockerfile`
  - `.github/workflows/lint.yaml`

## 実装時の注意
- 空ファイルや削除済みファイルのような GitHub PR 特有のケースを意識する
- PR コメントは重複投稿しない実装になっているため、コメント生成ロジック変更時は既存判定も確認する
- `review_gihub.go` はファイル名の綴りが `gihub` ですが、既存ファイル名として扱う

## 確認の目安
- ロジック変更時は `make lint` と `go test ./...` を最低限実行する
- Docker Action の変更時は `make build` も確認する
