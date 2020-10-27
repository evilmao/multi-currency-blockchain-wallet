#!/bin/bash

main_folder=${1}
app_file=${2}

DEV_TOOLS_SET_ENV=/opt/rh/devtoolset-7/enable
[ -f ${DEV_TOOLS_SET_ENV}  ] && source ${DEV_TOOLS_SET_ENV}

if [ "${buildxmr}" = "1" ]
then
    tags='buildxmr'
fi

go install -tags=${tags} ../cmd/${main_folder}
go build -tags=${tags} -o ./bin/${app_file} ../cmd/${main_folder}

echo
