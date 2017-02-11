#!/bin/sh

# Build for all platforms and architectures.
# Creates zip/tarball releases in build/ folder.

set -e

go get github.com/mitchellh/gox
gox -output='builds/{{.Dir}}_{{.OS}}_{{.Arch}}/{{.Dir}}'

cd builds

rm -rf *.zip *.tar.gz

for i in *
do
  cp ../README.md $i/
  cp ../LICENSE $i/
  case $i in
    *windows*) zip -r $i.zip $i/ ;;
    *) tar vzcf $i.tar.gz $i/ ;;
  esac
done

