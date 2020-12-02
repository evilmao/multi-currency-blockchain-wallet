#!/bin/bash

service_name=${1}
app_file=${service_name}

mkdir -p ../services/"${service_name}"/bin/
cp ./bin/deposit ../services/"${service_name}"/bin/"${app_file}"

printf '{
  "service_name": "wallet-deposit-%s",
  "type": "go",
  "app_file": "%s",
  "log_file_name": "wallet-deposit-%s.log",
  "bootstrap_args": ""
}
' "${service_name}" "${app_file}" "${service_name}" > ../services/"${service_name}"/service_spec.json

echo "'${service_name}'" >> ../.circleci/all_services.sh

echo