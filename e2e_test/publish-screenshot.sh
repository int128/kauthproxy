#!/usr/bin/env bash
set -o pipefail
set -eux

wiki_dir="$(mktemp -d)"
screenshot_path="${GITHUB_REF}/screenshot.png"
comment_body="## e2e-test
![screenshot](https://github.com/${GITHUB_REPOSITORY}/wiki/${screenshot_path})"

# publish the screenshot
git clone --depth=1 "https://x:${GITHUB_TOKEN}@github.com/${GITHUB_REPOSITORY}.wiki.git" "$wiki_dir"
mkdir -p $(dirname "$wiki_dir/$screenshot_path")
cp output/screenshot.png "$wiki_dir/$screenshot_path"
git -C "$wiki_dir" commit -a -m "ci-publish-screenshot"
git -C "$wiki_dir" push origin HEAD

# comment it to the pull request
if [ "$GITHUB_HEAD_REF" ]; then
  gh pr comment --body "$comment_body"
fi
