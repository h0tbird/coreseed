# coreseed

Define and deploy CoreOS clusters.

#### Deploy 3 masters on Packet.net:
```
for i in 1 2 3; do
coreseed udata \
--ns1-api-key aabbccddeeaabbccddee \
--domain cell-1.ewr.demo.lan \
--hostname core-${i} \
--role master \
--etcd-token UQRfgWywmLJta7RtHf5AYyV2ZH1qgPNa \
--ca-cert path/to/ca/cert.pem |
coreseed run-packet \
--hostname core-${i} \
--os coreos_alpha \
--api-key aabbccddeeaabbccddeeaabbccddeeaa \
--billing hourly \
--project-id aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee \
--plan baremetal_0 \
--facility ewr1
done
```
