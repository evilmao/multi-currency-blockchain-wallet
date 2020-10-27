#!/bin/bash

main_folder=${1}
service_name=${2}
app_file=${service_name}
server_port=${3}

set -e
go install ../cmd/${main_folder}
go build -o ./bin/${app_file} ../cmd/${main_folder}

mkdir -p ../services/${service_name}/bin/
cp ./bin/${app_file} ../services/${service_name}/bin/${app_file}

printf '{
  "service_name": "wallet-deposit-%s",
  "type": "go",
  "app_file": "%s",
  "server_port": %s,
  "log_file_name": "wallet-deposit-%s.log",
  "bootstrap_args": ""
}
' ${service_name} ${app_file} ${server_port} ${service_name} > ../services/${service_name}/service_spec.json

echo "'${service_name}'" >> ../.circleci/all_services.sh

echo