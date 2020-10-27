#!/bin/bash

source_path=${1}
source_package=${2}
target_file=${3}/import.go
package_name=${4}

printf 'package %s

import (
' ${package_name} > ${target_file}

for folder in $(ls -d ${source_path%%/}/*/)
do
    folder=${folder%%/}
    folder=${folder##*/}
    echo "    _ \"${source_package%%/}/${folder}\"" >> ${target_file}
done

 echo ")" >> ${target_file}

 echo