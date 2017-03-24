---
title: Deploy on Packet
---

# Deploy on Packet.net

<br>

<div class="alert alert-danger" role="alert">
  <span class="glyphicon glyphicon-exclamation-sign" aria-hidden="true"></span> Work in progress... <br/>
</div>

```bash
#!/bin/bash

CLUSTER_ID='<unique-cluster-id>'
DOMAIN='<managed-public-domain>'
EC2_REGION='<ec2-region>'
DNS_PROVIDER='<ns1|r53>'
DNS_API_KEY='<dns-private-key>'
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
  --host-id ${i} \
  --domain ${DOMAIN} \
  --ec2-region ${EC2_REGION} \
  --dns-provider ${DNS_PROVIDER} \
  --dns-api-key ${DNS_API_KEY} \
  --etcd-token ${ETCD_TOKEN} |

  katoctl pkt run \
  --api-key ${API_KEY} \
  --host-name master-${i}.cell-1.ewr \
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
  --host-id ${i} \
  --domain ${DOMAIN} \
  --ec2-region ${EC2_REGION} \
  --dns-provider ${DNS_PROVIDER} \
  --dns-api-key ${DNS_API_KEY} \
  --etcd-token ${ETCD_TOKEN} |

  katoctl pkt run \
  --api-key ${API_KEY} \
  --host-name worker-${i}.cell-1.ewr \
  --project-id ${PROJECT_ID} \
  --plan baremetal_0 \
  --os ${COREOS_CHANNEL} \
  --facility ewr1 \
  --billing hourly

done

#-------------------------
# Deploy one border node:
#-------------------------

for i in $(seq 1); do

  katoctl udata \
  --role border \
  --cluster-id ${CLUSTER_ID} \
  --master-count 3 \
  --host-id ${i} \
  --domain ${DOMAIN} \
  --ec2-region ${EC2_REGION} \
  --dns-provider ${DNS_PROVIDER} \
  --dns-api-key ${DNS_API_KEY} \
  --ca-cert-path ${CA_CERT_PATH} \
  --etcd-token ${ETCD_TOKEN} |

  katoctl pkt run \
  --api-key ${API_KEY} \
  --host-name border-${i}.cell-1.ewr \
  --project-id ${PROJECT_ID} \
  --plan baremetal_0 \
  --os ${COREOS_CHANNEL} \
  --facility ewr1 \
  --billing hourly

done
```
