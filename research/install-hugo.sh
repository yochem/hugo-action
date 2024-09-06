#!/usr/bin/env bash

os="$(go env GOOS)"
arch="$(go env GOARCH)"

[ "$os" = "darwin" ] && arch="universal"

TAG=latest

if [ "$TAG" = "latest" ]; then
	formatted_tag=""
else
	formatted_tag=$(echo "$TAG" | sed 's/^[^v].*/v&/')
fi

gh release download \
	--repo gohugoio/hugo \
	$formatted_tag \
	--pattern "*$os-$arch*" \
	--dir dl --clobber

tar -xzf dl/*
cp dl/hugo .
