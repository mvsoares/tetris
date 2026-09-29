# Homebrew Formula for Tetris
class Tetris < Formula
  desc "Terminal Tetris in Go with high-performance AI and multi-language support"
  homepage "https://github.com/mvsoares/tetris"
  version "1.0.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/mvsoares/tetris/releases/download/v1.0.0/tetris_1.0.0_darwin_arm64.tar.gz"
      # sha256 can be updated after release
    else
      url "https://github.com/mvsoares/tetris/releases/download/v1.0.0/tetris_1.0.0_darwin_amd64.tar.gz"
      # sha256 can be updated after release
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/mvsoares/tetris/releases/download/v1.0.0/tetris_1.0.0_linux_arm64.tar.gz"
    else
      url "https://github.com/mvsoares/tetris/releases/download/v1.0.0/tetris_1.0.0_linux_amd64.tar.gz"
    end
  end

  def install
    bin.install "tetris"
    (share/"tetris/models").install "models/move-risk.json" if File.exist?("models/move-risk.json")
  end

  test do
    assert_match "tetris v#{version}", shell_output("#{bin}/tetris --version")
  end
end
