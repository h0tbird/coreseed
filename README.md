# coreseed

Define and deploy CoreOS clusters.

```
coreseed data \
--hostname core-1 \
--domain demo.lan \
--role master \
--ns1-api-key aabbccddeeaabbccddee \
--ca-cert path/to/ca/cert.pem
```

```
coreseed run \
--api-key aabbccddeeaabbccddeeaabbccddeeaa \
--project-id aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee \
--hostname core-1 \
--plan baremetal_0 \
--os coreos_alpha \
--facility ewr1 \
--billing hourly
```
