# coreseed

Define and deploy CoreOS clusters.

#### Request an `etcd` bootstrapping token:
```
https://discovery.etcd.io/new?size=3
```

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

#### Deploy 3 masters on Amazon EC2:
```
for i in 1 2 3; do
coreseed udata \
--ns1-api-key xxxxxxxxxxxxxxxxxxxx \
--domain cell-1.ewr.demo.lan \
--hostname core-${i} \
--role master \
--etcd-token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
--ca-cert path/to/ca/cert.pem |
gzip --best | coreseed run-ec2 \
--region eu-west-1 \
--image-id ami-95bb00e6 \
--key-pair marc \
--instance-type t2.micro \
--vpc-id vpc-xxxxxxxx \
--subnet-id subnet-xxxxxxxx \
--sec-group-ids sg-xxx,sg-yyy \
done
```
