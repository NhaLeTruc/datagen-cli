# Homebrew formula for datagen-cli
# This formula is used to install datagen via Homebrew
#
# Usage:
#   brew install NhaLeTruc/tap/datagen
#
# For development:
#   brew install --build-from-source homebrew/datagen.rb

class Datagen < Formula
  desc "Generate PostgreSQL dump files from JSON schemas"
  homepage "https://github.com/NhaLeTruc/datagen-cli"
  version "1.0.0"
  license "MIT"

  # NOTE: Update these URLs when creating actual releases
  # The formula generator will populate these from GitHub releases
  if OS.mac? && Hardware::CPU.intel?
    url "https://github.com/NhaLeTruc/datagen-cli/releases/download/v1.0.0/datagen-v1.0.0-darwin-amd64.tar.gz"
    sha256 "PLACEHOLDER_SHA256_DARWIN_AMD64"
  elsif OS.mac? && Hardware::CPU.arm?
    url "https://github.com/NhaLeTruc/datagen-cli/releases/download/v1.0.0/datagen-v1.0.0-darwin-arm64.tar.gz"
    sha256 "PLACEHOLDER_SHA256_DARWIN_ARM64"
  elsif OS.linux? && Hardware::CPU.intel?
    url "https://github.com/NhaLeTruc/datagen-cli/releases/download/v1.0.0/datagen-v1.0.0-linux-amd64.tar.gz"
    sha256 "PLACEHOLDER_SHA256_LINUX_AMD64"
  elsif OS.linux? && Hardware::CPU.arm?
    url "https://github.com/NhaLeTruc/datagen-cli/releases/download/v1.0.0/datagen-v1.0.0-linux-arm64.tar.gz"
    sha256 "PLACEHOLDER_SHA256_LINUX_ARM64"
  end

  # Dependencies
  # datagen has no runtime dependencies (static binary)

  def install
    # Install the binary
    bin.install "datagen"

    # Install man pages
    man1.install Dir["docs/man/*.1"]

    # Install documentation
    doc.install "README.md" if File.exist?("README.md")
    doc.install "LICENSE" if File.exist?("LICENSE")

    # Install example schemas
    if File.directory?("docs/examples")
      (share/"datagen/examples").install Dir["docs/examples/*"]
    end

    # Generate shell completions
    generate_completions_from_executable(bin/"datagen", "completion")
  end

  test do
    # Test that the binary runs and shows version
    assert_match "datagen version", shell_output("#{bin}/datagen version")

    # Test basic schema validation
    (testpath/"test_schema.json").write <<~JSON
      {
        "version": "1.0",
        "database": {
          "name": "test_db"
        },
        "tables": {
          "users": {
            "columns": [
              {
                "name": "id",
                "type": "serial",
                "nullable": false
              },
              {
                "name": "email",
                "type": "varchar(255)",
                "nullable": false
              }
            ],
            "primary_key": ["id"],
            "row_count": 10
          }
        }
      }
    JSON

    # Validate the schema
    assert_match "Schema is valid", shell_output("#{bin}/datagen validate --input #{testpath}/test_schema.json")

    # Generate data to file
    system bin/"datagen", "generate", "--input", "#{testpath}/test_schema.json",
           "--output", "#{testpath}/output.sql", "--format", "sql"
    assert_predicate testpath/"output.sql", :exist?

    # Check that output contains expected SQL
    output_content = (testpath/"output.sql").read
    assert_match "CREATE TABLE users", output_content
    assert_match "INSERT INTO users", output_content
  end
end
