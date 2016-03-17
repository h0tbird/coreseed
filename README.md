# coreseed

Define and deploy CoreOS clusters.

#### Request an `etcd` bootstrapping token:
```
ETCD_TOKEN=$(curl -s https://discovery.etcd.io/new?size=3 | awk -F '/' '{print $NF}')
```

#### Deploy on Packet.net

##### 3 masters:
```
for i in 1 2 3; do
coreseed udata \
--ns1-api-key xxxxxxxxxxxxxxxxxxx \
--domain cell-1.ewr.demo.lan \
--hostname core-${i} \
--role master \
--etcd-token ${ETCD_TOKEN} \
--ca-cert path/to/ca/cert.pem | \
coreseed run-packet \
--hostname core-${i} \
--os coreos_alpha \
--api-key xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
--billing hourly \
--project-id xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \
--plan baremetal_0 \
--facility ewr1
done
```

##### 3 nodes:
```
todo
```

#### Deploy on Amazon EC2
```bash
#!/bin/bash

case $1 in

  #-------------------
  # Deploy 3 masters:
  #-------------------

  "masters")

    for i in 1 2 3; do

      coreseed udata \
      --hostname core-${i} \
      --domain cell-1.dc-1.demo.com \
      --role master \
      --ns1-api-key xxx \
      --ca-cert path/to/cert.pem \
      --etcd-token $(curl -s https://discovery.etcd.io/new?size=3 | awk -F '/' '{print $NF}') |

      gzip --best | coreseed run-ec2 \
      --region eu-west-1 \
      --image-id ami-95bb00e6 \
      --instance-type t2.medium \
      --key-pair xxx \
      --vpc-id vpc-xxx \
      --subnet-id subnet-xxx \
      --sec-group-ids sg-xxx,sg-xxx

    done ;;

  #-----------------
  # Deploy 3 nodes:
  #-----------------

  "nodes")

    for i in 4 5 6; do

      coreseed udata \
      --hostname core-${i} \
      --domain cell-1.dc-1.demo.com \
      --role node \
      --ns1-api-key xxx \
      --ca-cert path/to/cert.pem \
      --flannel-network 10.128.0.0/21 \
      --flannel-subnet-len 27 \
      --flannel-subnet-min 10.128.0.192 \
      --flannel-subnet-max 10.128.7.224 \
      --flannel-backend vxlan |

      gzip --best | coreseed run-ec2 \
      --region eu-west-1 \
      --image-id ami-95bb00e6 \
      --instance-type t2.medium \
      --key-pair xxx \
      --vpc-id vpc-xxx \
      --subnet-id subnet-xxx \
      --sec-group-ids sg-xxx,sg-xxx

    done ;;

esac
```

#### Deploy the full stack:
```
cd /etc/fleet
fleetctl submit zookeeper\@.service
fleetctl start zookeeper@1 zookeeper@2 zookeeper@3
fleetctl start mesos-master.service mesos-dns.service
fleetctl start marathon.service cadvisor.service
fleetctl start ceph-mon.service dnsmasq.service
fleetctl start ceph-osd.service mesos-node.service
```
