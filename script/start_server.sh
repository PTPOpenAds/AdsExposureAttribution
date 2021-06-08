#!/bin/sh

set -e

mkdir -p log_guest log_host data/guest_match_result data/host_match_result


nohup ./contract_attribution_server \
  --default_redis_config="{\"address\":[\"127.0.0.1:6379\"]}" \
  --role=GUEST \
  --server_address=:9080 \
  --metric_server_address=:8006 \
  --log_dir=./log_guest \
  --v=0 \
  --default_parallelism=50 \
  --default_queue_size=50 \
  --default_nlines=10 \
  --conv_encrypt_max_minute_freq=10000000 \
  --host_server_address="https://attribution.qq.com"
