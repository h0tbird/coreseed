# coreseed

Define and deploy CoreOS clusters.

#### Deploy 3 masters:
```
for i in 1 2 3; do
coreseed data \
--ns1-api-key aabbccddeeaabbccddee \
--domain cell-1.ewr.demo.lan \
--hostname core-${i} \
--role master \
--ca-cert path/to/ca/cert.pem |
coreseed run \
--hostname core-${i} \
--os coreos_alpha \
--api-key aabbccddeeaabbccddeeaabbccddeeaa \
--billing hourly \
--project-id aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee \
--plan baremetal_0 \
--facility ewr1
done
```
