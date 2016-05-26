### Deploy on Packet.net
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

      katoctl run-packet \
      --api-key xxx \
      --hostname master-${i}.cell-1.dc-1 \
      --project-id xxx \
      --plan baremetal_0 \
      --os coreos_alpha \
      --facility ewr1 \
      --billing hourly

    done ;;

  "workers")

    for i in $(seq $2); do

      katoctl udata \
      --role worker \
      --hostid ${i} \
      --domain cell-1.dc-1.demo.com \
      --ns1-api-key xxx \
      --ca-cert path/to/cert.pem \
      --flannel-network 10.128.0.0/21 \
      --flannel-subnet-len 27 \
      --flannel-subnet-min 10.128.0.192 \
      --flannel-subnet-max 10.128.7.224 \
      --flannel-backend vxlan |

      katoctl run-packet \
      --api-key xxx \
      --hostname worker-${i}.cell-1.dc-1 \
      --project-id xxx \
      --plan baremetal_0 \
      --os coreos_alpha \
      --facility ewr1 \
      --billing hourly

    done ;;

esac
```
