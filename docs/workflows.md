# CI/CD セットアップガイド

## 概要

このプロジェクトでは、`scripts/release.sh`とGitHub Actionsを組み合わせて、Docker imageのビルド・プッシュとHelmチャートリリースを自動化している。
リリース時には`helm/chart-releaser-action`を使用してHelmチャートの自動リリースを行う。

## 実行フロー
```bash
./scripts/release.sh 1.x.x
```

### 処理

#### Docker Job

**詳細ステップ**:
1. **code checkout** (`actions/checkout@v4`)
2. **バージョン抽出**: `v1.0.0` → `1.x.x`
3. **Docker Buildx setting**
4. **GHCR認証**
5. **タグ自動生成**: 
   - `ghcr.io/v01d42/nvidia-gpu-list-exporter:1.x.x`
   - `ghcr.io/v01d42/nvidia-gpu-list-exporter:latest`
6. **ビルド・プッシュ**

#### Helm Job

**詳細ステップ**:
1. **code checkout**
2. **Git setting**
3. **Latest Helm Chart install**
4. **version validation check**
5. **chart-releaser-action**:
   - チャート検証 (`helm lint`)
   - パッケージ化 (`helm package`)
   - GitHub Release自動作成
   - チャートファイル添付
   - gh-pagesブランチ更新
   - `index.yaml`自動生成

### 生成物

#### Dockerイメージ（GHCR）
- **Repo**: `ghcr.io/v01d42/nvidia-gpu-list-exporter`
- **Tag**: `1.x.x`, `latest`

#### Helmチャート（GitHub Pages）
- **Repo URL**: `https://V01d42.github.io/nvidia-gpu-list-exporter`
- **Chart File**: `nvidia-gpu-list-exporter-1.x.x.tgz`

#### GitHub Releases
- **Release**: `nvidia-gpu-list-exporter-1.x.x`
- **Files**: Chart tgzファイル

## 使用方法

### Dockerイメージ使用

```bash
# 最新版
docker pull ghcr.io/v01d42/nvidia-gpu-list-exporter:latest

# 特定バージョン
docker pull ghcr.io/v01d42/nvidia-gpu-list-exporter:1.x.x

# 実行
docker run --rm ghcr.io/v01d42/nvidia-gpu-list-exporter:latest
```

### Helmチャート使用

```bash
# 1. リポジトリ追加（初回のみ）
helm repo add nvidia-gpu-exporter https://V01d42.github.io/nvidia-gpu-list-exporter

# 2. チャート確認
helm search repo nvidia-gpu-exporter
helm show chart nvidia-gpu-exporter/nvidia-gpu-list-exporter

# 3. インストール
helm install my-exporter nvidia-gpu-exporter/nvidia-gpu-list-exporter

# 4. 特定バージョン指定
helm install my-exporter nvidia-gpu-exporter/nvidia-gpu-list-exporter --version 1.x.x

# 5. アップグレード
helm repo update
helm upgrade my-exporter nvidia-gpu-exporter/nvidia-gpu-list-exporter
```

## 初期設定

### 1. GitHub Pages有効化
```
リポジトリ設定 → Settings → Pages
Source: Deploy from a branch
Branch: gh-pages / (root)
```

### 2. 権限設定
```
Settings → Actions → General → Workflow permissions
Read and write permissions
```

### 3. パッケージ公開設定
```
リポジトリ → Packages → 各パッケージ → Package settings
Public (Make public)
```