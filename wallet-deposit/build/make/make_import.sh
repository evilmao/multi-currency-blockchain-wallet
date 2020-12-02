#!/bin/bash

source_path=${1}
source_package=${2}
main_path=${3}
target_file=${main_path}/imports/import.go
package_name=${4}

# 导入package依赖包
printf 'package %s

import (
' "${package_name}" > "${target_file}"

# shellcheck disable=SC2045
# write import rpc
for folder in $(ls -d "${source_path%%/}"/*/)
do
    folder=${folder%%/}
    folder=${folder##*/}
    echo "    _ \"${source_package%%/}/${folder}\"" >> "${target_file}"
done

# shellcheck disable=SC2012
# write import runnable
# shellcheck disable=SC2196
for folder in $(ls -d "${main_path%%/}"/*/ |egrep -v "(deposit|imports)")
do
    folder=${folder%%/}
    folder=${folder##*/}
    echo "    _ \"${source_package%%/}/${folder}\"" >> "${target_file}"
done


 echo ")" >> "${target_file}"

 echo