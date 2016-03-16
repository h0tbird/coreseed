# coreseed

Define and deploy CoreOS clusters.

#### Request an `etcd` bootstrapping token:
```
ETCD_TOKEN=$(curl -s https://discovery.etcd.io/new?size=3 | awk -F '/' '{print $NF}')
```

#### Deploy on Packet.net

##### 3 masters:
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

##### 3 nodes:
```
todo
```

#### Deploy on Amazon EC2

##### 3 masters:
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
--key-pair foo \
--instance-type t2.medium \
--vpc-id vpc-xxxxxxxx \
--subnet-id subnet-xxxxxxxx \
--sec-group-ids sg-xxx,sg-yyy
done
```

##### 3 nodes:
```
for i in 4 5 6; do
coreseed udata \
--ns1-api-key xxxxxxxxxxxxxxxxxxxx \
--domain cell-1.ewr.demo.lan \
--hostname core-${i} \
--role node \
--ca-cert --ca-cert path/to/ca/cert.pem | \
gzip --best | coreseed run-ec2 \
--region eu-west-1 \
--image-id ami-95bb00e6 \
--key-pair foo \
--instance-type t2.medium \
--vpc-id vpc-xxxxxxxx \
--subnet-id subnet-xxxxxxxx \
--sec-group-ids sg-xxx,sg-yyy
done
```

#### Deploy the full stack:
```
cd /etc/fleet
fleetctl submit zookeeper\@.service
fleetctl start zookeeper@1 zookeeper@2 zookeeper@3
fleetctl start mesos-master.service mesos-dns.service
fleetctl start marathon.service cadvisor.service
fleetctl start ceph-mon.service dnsmasq.service
fleetctl start ceph-osd.service mesos-node.service
```
