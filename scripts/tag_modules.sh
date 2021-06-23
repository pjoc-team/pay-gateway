#!/usr/bin/env bash
version="$1"
if [ -z "$version" ]; then
  echo "version is empty!"
  exit 1
fi

find . -name "go.mod" | sed 's~\./~~' | sed "s/go.mod/${version}/" | while read t; do
  git tag $t
done
