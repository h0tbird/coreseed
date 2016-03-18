### Deploy on Packet.net
```bash
#!/bin/bash

case $1 in

  "masters")

    ETCD_TOKEN=$(curl -s https://discovery.etcd.io/new?size=3 | awk -F '/' '{print $NF}')

    for i in 1 2 3; do

      coreseed udata \
      --hostname core-${i} \
      --domain cell-1.dc-1.demo.com \
      --role master \
      --ns1-api-key xxx \
      --ca-cert path/to/cert.pem \
      --etcd-token ${ETCD_TOKEN} |

      coreseed run-packet \
      --api-key xxx \
      --hostname core-${i} \
      --project-id xxx \
      --plan baremetal_0 \
      --os coreos_alpha \
      --facility ewr1 \
      --billing hourly

    done ;;

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

      coreseed run-packet \
      --api-key xxx \
      --hostname core-${i} \
      --project-id xxx \
      --plan baremetal_0 \
      --os coreos_alpha \
      --facility ewr1 \
      --billing hourly

    done ;;

esac
```
