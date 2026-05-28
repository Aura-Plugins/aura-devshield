# Homebrew formula for Aura DevShield.
#
# This is a binary formula — it installs pre-built release binaries rather
# than compiling from source. No Go toolchain or Xcode CLT required.
#
# To publish this formula, copy it to the 'homebrew-tap' repository under
# the Aura-Plugins GitHub organisation. Users install via:
#
#   brew tap aura-plugins/tap
#   brew install aura-devshield
#
# On every release:
#   1. Update `version` to the new tag (without the leading "v").
#   2. Update the sha256 values below. Get them from dist/checksums.txt
#      produced by `make checksums`, or from the GitHub Release page.
#
class AuraDevshield < Formula
  desc "Local-first developer supply-chain security visibility tool"
  homepage "https://github.com/Aura-Plugins/aura-devshield"
  version "0.3.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Aura-Plugins/aura-devshield/releases/download/v#{version}/aura-devshield-darwin-arm64"
      sha256 "REPLACE_WITH_SHA256_FROM_CHECKSUMS_TXT"
    end

    on_intel do
      url "https://github.com/Aura-Plugins/aura-devshield/releases/download/v#{version}/aura-devshield-darwin-amd64"
      sha256 "REPLACE_WITH_SHA256_FROM_CHECKSUMS_TXT"
    end
  end

  on_linux do
    on_amd64 do
      url "https://github.com/Aura-Plugins/aura-devshield/releases/download/v#{version}/aura-devshield-linux-amd64"
      sha256 "REPLACE_WITH_SHA256_FROM_CHECKSUMS_TXT"
    end
  end

  def install
    bin.install Dir["aura-devshield-*"].first => "aura-devshield"
  end

  test do
    assert_match "aura-devshield #{version}", shell_output("#{bin}/aura-devshield --version")
  end
end
