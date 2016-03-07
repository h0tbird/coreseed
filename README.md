# coreseed

Define and deploy CoreOS clusters.

#### Etcd clustering:
##### Discovery
Generate an etcd discovery token:
```
curl 'https://discovery.etcd.io/new?size=3'
```
##### Static
If you don't provide an `etcd` discovery token:
- The cluster will attempt to start statically with `initial-cluster`...
- `core-1=http://core-1:2380,core-2=http://core-2:2380,core-3=http://core-3:2380`
- For this to work DNS must resolve `core-1`, `core-2` and `core-3`.

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
