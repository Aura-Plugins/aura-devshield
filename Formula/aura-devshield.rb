# Homebrew formula for Aura DevShield.
#
# To publish this formula, create a tap repository named 'homebrew-tap' under
# the same GitHub account, place this file there, then users install via:
#
#   brew tap aura-plugins/tap
#   brew install aura-devshield
#
# Update the `url` and `sha256` on every release. The SHA256 of the source
# tarball is printed by GitHub when you create the release, or compute it with:
#
#   curl -sSfL <tarball_url> | sha256sum
#
class AuraDevshield < Formula
  desc "Local-first developer supply-chain security visibility tool"
  homepage "https://github.com/Aura-Plugins/aura-devshield"
  url "https://github.com/Aura-Plugins/aura-devshield/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "cbe7dbba10f10f729b52206ef6076c4d5ce8622a33543310780280400dadd64e"
  license "MIT"
  head "https://github.com/Aura-Plugins/aura-devshield.git", branch: "main"

  depends_on "go" => :build

  def install
    ldflags = "-s -w -X main.version=#{version}"
    system "go", "build", *std_go_args(ldflags: ldflags), "./cmd/aura-devshield"
  end

  test do
    # Scan exits 0 even with no extensions directory (handled gracefully).
    assert_match "aura-devshield", shell_output("#{bin}/aura-devshield --version")
  end
end
