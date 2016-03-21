#### Deploy on Amazon EC2
```bash
#!/bin/bash

case $1 in

  "masters")

    ETCD_TOKEN=$(curl -s https://discovery.etcd.io/new?size=3 | awk -F '/' '{print $NF}')

    for i in $(seq $2); do

      katoctl udata \
      --role master \
      --hostid ${i} \
      --domain cell-1.dc-1.demo.com \
      --ns1-api-key xxx \
      --ca-cert path/to/cert.pem \
      --etcd-token ${ETCD_TOKEN} |

      gzip --best | katoctl run-ec2 \
      --hostname master-${i}.cell-1.dc-1 \
      --region eu-west-1 \
      --image-id ami-7b971208 \
      --instance-type t2.medium \
      --key-pair xxx \
      --vpc-id vpc-xxx \
      --subnet-id subnet-xxx \
      --sec-group-ids sg-xxx,sg-xxx

    done ;;

  "nodes")

    for i in $(seq $2); do

      katoctl udata \
      --role node \
      --hostid ${i} \
      --domain cell-1.dc-1.demo.com \
      --ns1-api-key xxx \
      --ca-cert path/to/cert.pem \
      --flannel-network 10.128.0.0/21 \
      --flannel-subnet-len 27 \
      --flannel-subnet-min 10.128.0.192 \
      --flannel-subnet-max 10.128.7.224 \
      --flannel-backend vxlan |

      gzip --best | katoctl run-ec2 \
      --hostname node-${i}.cell-1.dc-1 \
      --region eu-west-1 \
      --image-id ami-7b971208 \
      --instance-type t2.medium \
      --key-pair xxx \
      --vpc-id vpc-xxx \
      --subnet-id subnet-xxx \
      --sec-group-ids sg-xxx,sg-xxx

    done ;;

esac
```
