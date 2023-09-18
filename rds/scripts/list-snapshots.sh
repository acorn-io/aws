#!/bin/bash
set -x
aws rds describe-db-cluster-snapshots --db-cluster-identifier "${DB_CLUSTER_ID}" --query "DBClusterSnapshots[*].DBClusterSnapshotArn" | jq -r '.[]' > /run/secrets/output