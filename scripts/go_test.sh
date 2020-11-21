#!/usr/bin/env bash

find . -name "*.go" | while read f; do
    echo "${f%/*}"
done | sort | uniq | while read f; do
    go test -race $f
    go test -v -cpuprofile="$f.cpu.prof" -memprofile "$f.mem.prof" -bench -x "$f"
    go tool pprof -svg "$f.cpu.prof" > $f.prof.svg
done
