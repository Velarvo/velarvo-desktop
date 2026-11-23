#!/usr/bin/env bash
set -euo pipefail

version="${1#v}"
if ! [[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+([-+][0-9A-Za-z.-]+)?$ ]]; then
  echo "usage: $0 <semver>  (got '${1:-}')" >&2
  exit 1
fi

cd "$(dirname "$0")/.."
export VERSION="$version"
export NUMERIC_VERSION="${version%%[-+]*}"

perl -pi -e 's/^(  version: )"[^"]*"/$1"$ENV{VERSION}"/' build/config.yml

for plist in build/darwin/Info.plist build/darwin/Info.dev.plist; do
  perl -0pi -e 's{(<key>CFBundleShortVersionString</key>\s*<string>)[^<]*(</string>)}{$1$ENV{VERSION}$2}g; s{(<key>CFBundleVersion</key>\s*<string>)[^<]*(</string>)}{$1$ENV{NUMERIC_VERSION}$2}g' "$plist"
done

perl -pi -e 's/("(?:file_version|product_version)": ")[^"]*"/$1$ENV{NUMERIC_VERSION}"/g; s/("(?:FileVersion|ProductVersion)": ")[^"]*"/$1$ENV{VERSION}"/g' build/windows/info.json

perl -pi -e 's/(!define INFO_PRODUCTVERSION )"[^"]*"/$1"$ENV{NUMERIC_VERSION}"/' build/windows/nsis/wails_tools.nsh

perl -pi -e 's/^version: "[^"]*"/version: "$ENV{VERSION}"/' build/linux/nfpm/nfpm.yaml

echo "Version set to $version"
