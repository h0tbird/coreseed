# coreseed

Define and deploy CoreOS clusters.

```
coreseed data \
--hostname core-1 \
--domain demo.lan \
--role slave \
--ns1-api-key aabbccddeeaabbccddee \
--tags osd
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
