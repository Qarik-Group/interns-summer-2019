#!/bin/sh

echo -e ""
docker exec -i etcd-cluster_etcd2_1 /bin/sh << EOF
echo "===========putting the key-value in etcd==========="
etcdctl put newkey "Test for the new plugin"
echo "===========getting key-value out of etcd==========="
etcdctl get newkey
echo "===========printing out member list==========="
etcdctl --write-out=table member list
echo "===========printing out member health==========="
etcdctl endpoint health
echo -e "===========exiting etcd===========\n"
exit
EOF
