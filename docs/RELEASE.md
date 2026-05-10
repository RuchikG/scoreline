# Release Checklist

## One-time setup

1. Create a public GitHub repository named `RuchikG/homebrew-tap`.
2. Create a fine-grained GitHub token with contents read/write access to `RuchikG/homebrew-tap`.
3. Add that token to this repository as an Actions secret named `TAP_GITHUB_TOKEN`.
4. Ensure this repository's Actions settings allow `GITHUB_TOKEN` to create releases with write permissions.

## Release

1. Make sure `CHANGELOG.md` has the release notes moved from `Unreleased` into a versioned section.
2. Run:

   ```bash
   go test ./...
   go build -trimpath ./...
   ```

3. Tag and push:

   ```bash
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```

4. Confirm the `Release` workflow publishes:
   - GitHub release assets for macOS, Linux, and Windows.
   - `checksums.txt`.
   - `Formula/scoreline.rb` in `RuchikG/homebrew-tap`.

5. Verify install paths:

   ```bash
   brew tap RuchikG/tap
   brew install scoreline
   scoreline --version
   ```

   ```bash
   curl -fsSL https://raw.githubusercontent.com/RuchikG/scoreline/main/scripts/install.sh | bash
   scoreline --version
   ```
