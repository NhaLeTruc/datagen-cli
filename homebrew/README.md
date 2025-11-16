# Homebrew Formula for datagen-cli

This directory contains the Homebrew formula for installing datagen-cli on macOS and Linux.

## Installation (for end users)

Once the tap is published, users can install datagen with:

```bash
# Tap the repository (first time only)
brew tap NhaLeTruc/tap

# Install datagen
brew install datagen

# Or install directly
brew install NhaLeTruc/tap/datagen
```

## Formula Development

### Structure

- `datagen.rb` - Main Homebrew formula
- `../scripts/update-homebrew.sh` - Script to update checksums after releases

### Creating a New Release

1. **Create release artifacts**:
   ```bash
   ./scripts/release.sh --version 1.0.0
   ```

2. **Update formula with checksums**:
   ```bash
   ./scripts/update-homebrew.sh --version 1.0.0
   ```

3. **Test the formula locally**:
   ```bash
   brew install --build-from-source homebrew/datagen.rb
   brew test datagen
   ```

4. **Commit and push**:
   ```bash
   git add homebrew/datagen.rb
   git commit -m "Update Homebrew formula for v1.0.0"
   git push
   ```

### Publishing to Homebrew

#### Option 1: Personal Tap (Recommended for Getting Started)

Create a separate tap repository:

```bash
# Create tap repository
git clone https://github.com/NhaLeTruc/homebrew-tap.git
cd homebrew-tap

# Copy formula
cp ../datagen-cli/homebrew/datagen.rb Formula/datagen.rb

# Commit and push
git add Formula/datagen.rb
git commit -m "Add datagen formula"
git push
```

Users can then install with:
```bash
brew tap NhaLeTruc/tap
brew install datagen
```

#### Option 2: Official Homebrew Core (For Established Projects)

Submit a pull request to [homebrew-core](https://github.com/Homebrew/homebrew-core):

1. Fork homebrew-core
2. Add `datagen.rb` to `Formula/` directory
3. Ensure formula passes `brew audit --strict --online datagen`
4. Submit PR with description

Requirements for homebrew-core:
- Stable, versioned releases
- Active maintenance
- Clear documentation
- Passing all tests

### Testing the Formula

```bash
# Install from local file
brew install --build-from-source homebrew/datagen.rb

# Run formula tests
brew test datagen

# Audit the formula
brew audit --strict --online datagen

# Check for common issues
brew style datagen
```

### Updating the Formula

When releasing a new version:

1. Update `version` in formula
2. Update download URLs for all platforms
3. Update SHA256 checksums (use update-homebrew.sh script)
4. Test locally
5. Commit and push

### Formula Components

**Header**:
- `desc`: One-line description
- `homepage`: Project homepage URL
- `version`: Current version
- `license`: Software license

**Platform-specific URLs**:
- macOS Intel (darwin-amd64)
- macOS Apple Silicon (darwin-arm64)
- Linux Intel (linux-amd64)
- Linux ARM (linux-arm64)

**Install section**:
- Binary installation
- Man page installation
- Documentation installation
- Example schemas installation
- Shell completion generation

**Test section**:
- Version check
- Schema validation
- Data generation
- SQL output verification

### Troubleshooting

**Formula audit fails**:
- Run `brew audit --strict --online datagen` to see specific issues
- Common issues: missing license, incorrect URLs, missing tests

**Installation fails**:
- Check SHA256 checksums match release artifacts
- Verify release URLs are accessible
- Test with `brew install --verbose --debug`

**Tests fail**:
- Run `brew test datagen --verbose` for detailed output
- Ensure all test files are included in release
- Check that binary has correct permissions

### Resources

- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Homebrew Formula Guidelines](https://docs.brew.sh/Acceptable-Formulae)
- [Creating Taps](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)
- [Formula Language Reference](https://rubydoc.brew.sh/Formula)

## Notes

- Formula supports macOS (Intel and Apple Silicon) and Linux (amd64 and arm64)
- No runtime dependencies (statically compiled Go binary)
- Man pages automatically installed to `$(brew --prefix)/share/man/man1/`
- Examples installed to `$(brew --prefix)/share/datagen/examples/`
- Shell completions generated automatically
