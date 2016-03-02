# coreseed

Define and deploy CoreOS clusters.

```
coreseed data \
-h core-1 \
-d demo.lan \
-r slave \
-k aabbccddeeaabbccddee \
-t osd
```

```
coreseed run \
-k aabbccddeeaabbccddeeaabbccddeeaa \
-i aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee \
-h core-1 \
-p baremetal_0 \
-o coreos_alpha \
-f ewr1 \
-b hourly
```
