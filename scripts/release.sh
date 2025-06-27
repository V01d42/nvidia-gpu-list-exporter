#!/bin/bash

set -e  # エラー時に停止

# 色付きメッセージ用の定数
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ヘルプメッセージ
show_help() {
    echo "Usage: $0 <version>"
    echo ""
    echo "Example:"
    echo "  $0 1.0.2"
    echo ""
    echo "This script will:"
    echo "  1. Update Helm chart versions"
    echo "  2. Commit changes"
    echo "  3. Create and push git tag"
    echo "  4. Push changes to main branch"
    echo ""
}

# バージョン形式の検証
validate_version() {
    if [[ ! "$1" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo -e "${RED}Error: Version must be in format x.y.z (e.g., 1.0.2)${NC}"
        exit 1
    fi
}

# Gitの状態確認
check_git_status() {
    # 現在のブランチがmainかチェック
    current_branch=$(git symbolic-ref --short HEAD)
    if [ "$current_branch" != "main" ]; then
        echo -e "${RED}Error: You must be on 'main' branch. Current branch: $current_branch${NC}"
        exit 1
    fi
    
    # 未コミットの変更がないかチェック
    if ! git diff-index --quiet HEAD --; then
        echo -e "${RED}Error: You have uncommitted changes. Please commit or stash them first.${NC}"
        git status --short
        exit 1
    fi
    
    # リモートと同期しているかチェック
    echo -e "${BLUE}Fetching latest changes...${NC}"
    git fetch origin
    
    if [ $(git rev-list HEAD...origin/main --count) -ne 0 ]; then
        echo -e "${RED}Error: Your branch is not up to date with origin/main${NC}"
        echo "Please run: git pull origin main"
        exit 1
    fi
}

# ファイル更新
update_files() {
    local version=$1
    
    echo -e "${BLUE}Updating Helm chart files...${NC}"
    
    # Chart.yaml更新
    sed -i "s/version: .*/version: $version/" ./helm/nvidia-gpu-list-exporter/Chart.yaml
    sed -i "s/appVersion: .*/appVersion: $version/" ./helm/nvidia-gpu-list-exporter/Chart.yaml
    
    # values.yaml更新
    sed -i "s/tag: .*/tag: \"$version\"/" ./helm/nvidia-gpu-list-exporter/values.yaml
    
    echo -e "${GREEN}Updated files:${NC}"
    echo "  - helm/nvidia-gpu-list-exporter/Chart.yaml"
    echo "  - helm/nvidia-gpu-list-exporter/values.yaml"
}

# 変更の確認
show_changes() {
    echo -e "\n${YELLOW}Changes to be committed:${NC}"
    git diff --color=always ./helm/nvidia-gpu-list-exporter/Chart.yaml ./helm/nvidia-gpu-list-exporter/values.yaml
    
    echo -e "\n${YELLOW}Do you want to proceed with these changes? (y/N)${NC}"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        echo -e "${RED}Release canceled${NC}"
        # 変更を元に戻す
        git checkout -- ./helm/nvidia-gpu-list-exporter/Chart.yaml ./helm/nvidia-gpu-list-exporter/values.yaml
        exit 1
    fi
}

# コミットとプッシュ
commit_and_push() {
    local version=$1
    
    echo -e "${BLUE}Committing changes...${NC}"
    git add ./helm/nvidia-gpu-list-exporter/Chart.yaml ./helm/nvidia-gpu-list-exporter/values.yaml
    git commit -m "chore: release v$version"
    
    echo -e "${BLUE}Creating tag v$version...${NC}"
    git tag "v$version"
    
    echo -e "${BLUE}Pushing to origin...${NC}"
    git push origin main
    git push origin "v$version"
    
    echo -e "${GREEN}Successfully released v$version!${NC}"
    echo -e "${GREEN}Tag v$version has been pushed and GitHub Actions will build/deploy automatically${NC}"
}

# メイン処理
main() {
    # 引数チェック
    if [ $# -eq 0 ] || [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
        show_help
        exit 0
    fi
    
    local version=$1
    
    echo -e "${BLUE}Starting release process for version $version${NC}"
    
    # 各種チェック
    validate_version "$version"
    check_git_status
    
    # ファイル更新
    update_files "$version"
    
    # 変更確認
    show_changes
    
    # コミットとプッシュ
    commit_and_push "$version"
}

# スクリプト実行
main "$@" 