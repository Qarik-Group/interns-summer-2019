#!/bin/bash
echo "Logging into the etcd sluster"
kubectl exec -it etcd0 sh << EOF
export ETCDCTL_API=3
etcdctl --write-out=table member list
etcdctl endpoint health
etcdctl put /test newkey "Test for the new plugin"
etcdctl get /test newkey
EOF
