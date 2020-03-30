class TerragruntAtlantisConfig < Formula
  desc "Generates Atlantis Config for Terragrunt projects"
  homepage "https://github.com/transcend-io/terragrunt-atlantis-config"
  url "https://s3-eu-west-1.amazonaws.com/downloads.heft.io/0.0.3/heft_0.0.3_darwin_amd64.zip"
  url "https://homebrew.transcend.io/terragrunt-atlantis-config/0.0.1/terragrunt-atlantis-config_0.0.1_darwin_amd64.zip"
  version "0.0.1"
  sha256 "6c7a37aa078243d4df48d9bfeb9df4ff5f97f06e442aeaf78589d927e51fc85b"

  def install
    bin.install "terragrunt-atlantis-config"
  end
end