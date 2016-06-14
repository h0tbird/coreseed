### Deploy on Packet.net
```bash
#!/bin/bash

CLUSTER_ID='<unique-cluster-id>'
DOMAIN='<ns1-managed-public-domain>'
EC2_REGION='<ec2-region>'
NS1_API_KEY='<ns1-private-key>'
ETCD_TOKEN=$(curl -s https://discovery.etcd.io/new?size=3 | awk -F '/' '{print $NF}')
API_KEY='<packet-private-api-key>'
PROJECT_ID='<packet-project-id>'
COREOS_CHANNEL='coreos_stable'

#----------------------------
# Deploy three master nodes:
#----------------------------

for i in $(seq 3); do

  katoctl udata \
  --role master \
  --cluster-id ${CLUSTER_ID} \
  --master-count 3 \
  --hostid ${i} \
  --domain ${DOMAIN} \
  --ec2-region ${EC2_REGION} \
  --ns1-api-key ${NS1_API_KEY} \
  --etcd-token ${ETCD_TOKEN} |

  katoctl pkt run \
  --api-key ${API_KEY} \
  --hostname master-${i}.cell-1.ewr \
  --project-id ${PROJECT_ID} \
  --plan baremetal_0 \
  --os ${COREOS_CHANNEL} \
  --facility ewr1 \
  --billing hourly

done

#-------------------------
# Deploy two slave nodes:
#-------------------------

for i in $(seq 2); do

  katoctl udata \
  --role worker \
  --cluster-id ${CLUSTER_ID} \
  --master-count 3 \
  --hostid ${i} \
  --domain ${DOMAIN} \
  --ec2-region ${EC2_REGION} \
  --ns1-api-key ${NS1_API_KEY} \
  --etcd-token ${ETCD_TOKEN} |

  katoctl pkt run \
  --api-key ${API_KEY} \
  --hostname worker-${i}.cell-1.ewr \
  --project-id ${PROJECT_ID} \
  --plan baremetal_0 \
  --os ${COREOS_CHANNEL} \
  --facility ewr1 \
  --billing hourly

done

#-----------------------
# Deploy one edge node:
#-----------------------

for i in $(seq 1); do

  katoctl udata \
  --role edge \
  --cluster-id ${CLUSTER_ID} \
  --master-count 3 \
  --hostid ${i} \
  --domain ${DOMAIN} \
  --ec2-region ${EC2_REGION} \
  --ns1-api-key ${NS1_API_KEY} \
  --ca-cert ${CA_CERT} \
  --etcd-token ${ETCD_TOKEN} |

  katoctl pkt run \
  --api-key ${API_KEY} \
  --hostname edge-${i}.cell-1.ewr \
  --project-id ${PROJECT_ID} \
  --plan baremetal_0 \
  --os ${COREOS_CHANNEL} \
  --facility ewr1 \
  --billing hourly

done
```
