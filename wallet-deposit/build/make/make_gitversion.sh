#!/bin/bash

target_file=${1}/version.go
package_name=${2}

# 获取最新代码库版本hash
hash=$(git rev-parse --short HEAD)

function joins { echo "$*"; }
# shellcheck disable=SC2046
tag=$(joins $(git tag --points-at "${hash}"))

# 修改代码 version.go 文件
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
