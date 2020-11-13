#!/bin/bash

main_folder=${1}
service_name=${2}

DEV_TOOLS_SET_ENV=/opt/rh/devtoolset-7/enable
# shellcheck disable=SC1090
[ -f ${DEV_TOOLS_SET_ENV}  ] && source ${DEV_TOOLS_SET_ENV}

if [ "${buildxmr}" = "1" ]
then
    tags='buildxmr'
fi

set -e
go install -tags=${tags} ../cmd/"${main_folder}"
go build -tags=${tags} -o ./bin/"${service_name}" ../cmd/"${main_folder}"/main.go

echo