# coreseed

Define and deploy CoreOS clusters.

#### Request an `etcd` bootstrapping token:
```
ETCD_TOKEN=$(curl -s https://discovery.etcd.io/new?size=3 | awk -F '/' '{print $NF}')
```

#### Deploy 3 masters on Packet.net:
```
for i in 1 2 3; do
coreseed udata \
--ns1-api-key xxxxxxxxxxxxxxxxxxx \
--domain cell-1.ewr.demo.lan \
--hostname core-${i} \
--role master \
--etcd-token ${ETCD_TOKEN} \
--ca-cert path/to/ca/cert.pem |
coreseed run-packet \
--hostname core-${i} \
--os coreos_alpha \
--api-key xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
--billing hourly \
--project-id xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \
--plan baremetal_0 \
--facility ewr1
done
```

#### Deploy 3 masters on Amazon EC2:
```
for i in 1 2 3; do
coreseed udata \
--ns1-api-key xxxxxxxxxxxxxxxxxxxx \
--domain cell-1.ewr.demo.lan \
--hostname core-${i} \
--role master \
--etcd-token ${ETCD_TOKEN} \
--ca-cert path/to/ca/cert.pem | \
gzip --best | coreseed run-ec2 \
--region eu-west-1 \
--image-id ami-95bb00e6 \
--key-pair marc \
--instance-type t2.micro \
--vpc-id vpc-xxxxxxxx \
--subnet-id subnet-xxxxxxxx \
--sec-group-ids sg-xxx,sg-yyy
done
```
