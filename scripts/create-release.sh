#!/bin/bash

# 创建 GitHub Release 的脚本
# 使用方法: ./scripts/create-release.sh v0.1.5

set -e

VERSION=$1

if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v0.1.5"
    exit 1
fi

# 检查是否有未提交的更改
if ! git diff-index --quiet HEAD --; then
    echo "Error: You have uncommitted changes. Please commit or stash them first."
    exit 1
fi

# 检查标签是否已存在
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo "Tag $VERSION already exists. Do you want to delete and recreate it? (y/n)"
    read -r answer
    if [ "$answer" = "y" ]; then
        git tag -d "$VERSION"
        git push origin --delete "$VERSION" 2>/dev/null || true
    else
        echo "Aborted."
        exit 1
    fi
fi

# 获取最新的提交信息
LATEST_COMMIT=$(git log -1 --pretty=format:"%h %s")

# 创建标签
echo "Creating tag $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION

$LATEST_COMMIT

For full changelog, see: https://github.com/Alfonsxh/gitlab-merge-alert-go/compare/$(git describe --tags --abbrev=0 2>/dev/null || echo 'initial')...$VERSION"

# 推送标签
echo "Pushing tag $VERSION to GitHub..."
git push origin "$VERSION"

echo "✅ Tag $VERSION created and pushed successfully!"
echo "🚀 GitHub Actions will now build and create the release automatically."
echo "📦 Check the Actions tab for build progress: https://github.com/Alfonsxh/gitlab-merge-alert-go/actions"