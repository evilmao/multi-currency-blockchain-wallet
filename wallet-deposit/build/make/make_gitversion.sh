#!/bin/bash

target_file=${1}/version.go
package_name=${2}

hash=$(git rev-parse --short HEAD)

function joins { echo "$*"; }
tag=$(joins $(git tag --points-at ${hash}))

printf 'package %s

const (
    VersionHash = "%s"
    VersionTag  = "%s"
)

func Version() string {
	return VersionTag + "(" + VersionHash + ")"
}
' "${package_name}" "${hash}" "${tag}" > "${target_file}"

echo
