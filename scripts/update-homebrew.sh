#!/bin/bash

# 在当前项目中更新Homebrew formula的脚本
set -e

VERSION=${1:-$(git describe --tags --abbrev=0)}
PROJECT_REPO="zzxwill/aigit"

echo "🍺 Updating Homebrew formula for version $VERSION"

# 1. 下载并计算SHA256
echo "📦 Downloading release tarball..."
TARBALL_URL="https://github.com/$PROJECT_REPO/archive/refs/tags/$VERSION.tar.gz"
SHA256=$(curl -sL "$TARBALL_URL" | shasum -a 256 | cut -d' ' -f1)

echo "🔍 SHA256: $SHA256"

# 2. 确保Formula目录存在
mkdir -p Formula

# 3. 更新formula
echo "✏️  Updating formula..."
cat > "Formula/aigit.rb" << EOF
class Aigit < Formula
  desc "AI-powered Git commit message generator using LLM"
  homepage "https://github.com/$PROJECT_REPO"
  url "https://github.com/$PROJECT_REPO/archive/refs/tags/$VERSION.tar.gz"
  sha256 "$SHA256"
  license "Apache-2.0"
  head "https://github.com/$PROJECT_REPO.git", branch: "master"

  depends_on "go" => :build

  def install
    ldflags = %W[
      -s -w
      -X main.Version=#{version}
    ]

    system "go", "build", *std_go_args(ldflags: ldflags), "./main.go"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/aigit version")
    assert_match "Generate git commit message", shell_output("#{bin}/aigit help")
  end

  def caveats
    <<~EOS
      Before using aigit, configure an AI provider:
        aigit auth add openai YOUR_API_KEY
        aigit auth add gemini YOUR_API_KEY
        aigit auth add deepseek YOUR_API_KEY
        aigit auth add doubao YOUR_API_KEY YOUR_ENDPOINT_ID

      Then use: aigit commit
    EOS
  end
end
EOF

# 4. 提交更新到当前项目
echo "📤 Committing updates..."
git add Formula/aigit.rb

# 检查是否有变更需要提交
if git diff --staged --quiet; then
    echo "ℹ️  No changes to Formula, skipping commit"
else
    git commit -m "chore: update homebrew formula to $VERSION

- Update formula for version $VERSION
- SHA256: $SHA256"

    echo "✅ Homebrew formula updated successfully!"
fi

echo "🎉 Users can now install with:"
echo "   brew install https://raw.githubusercontent.com/$PROJECT_REPO/master/Formula/aigit.rb"
