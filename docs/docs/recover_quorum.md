---
title: Recover quorum nodes
---

# Recover quorum nodes

Let's destroy the quorum master and recreate it from scratch. I have 1 `border`, 3 `quorum`, 3 `master` and 3 `worker` nodes up and running on *EC2*. I am also connected to the cluster via *Pritunl* which is running on the `border` node. The cluster is running *The Voting App* which is a 5 container demo application deployed with *Marathon*. By destroying one `quorum` node I am affecting the `zookeeper` and `etcd2` services among others (full list below):

```
core@quorum-1 ~ $ katostat
LoadState=loaded  ActiveState=active  SubState=running  Id=etcd2.service
LoadState=loaded  ActiveState=active  SubState=running  Id=docker.service
LoadState=loaded  ActiveState=active  SubState=running  Id=zookeeper.service
LoadState=loaded  ActiveState=active  SubState=running  Id=rexray.service
LoadState=loaded  ActiveState=active  SubState=running  Id=cadvisor.service
LoadState=loaded  ActiveState=active  SubState=running  Id=node-exporter.service
LoadState=loaded  ActiveState=active  SubState=running  Id=zookeeper-exporter.service
LoadState=loaded  ActiveState=active  SubState=waiting  Id=etchost.timer
```

## 1. Who is the elected master?

The answer is `quorum-3` for *zookeeper*:

```
core@quorum-1 ~ $ loopssh quorum "echo srvr | ncat \$(hostname) 2181 | grep Mode"
--[ quorum-1.cell-1.dc-1.demo.lan ]--
Mode: follower
--[ quorum-2.cell-1.dc-1.demo.lan ]--
Mode: follower
--[ quorum-3.cell-1.dc-1.demo.lan ]--
Mode: leader
```

And conveniently `quorum-3` for *etcd* too:

```
core@quorum-1 ~ $ etcdctl member list
deff3e5ba168963: name=quorum-2 peerURLs=http://10.136.89.99:2380 clientURLs=http://10.136.89.99:2379 isLeader=false
498aec03aa857694: name=quorum-3 peerURLs=http://10.136.83.75:2380 clientURLs=http://10.136.83.75:2379 isLeader=true
fadca54d49882874: name=quorum-1 peerURLs=http://10.136.113.221:2380 clientURLs=http://10.136.113.221:2379 isLeader=false
```

This is the *ARP* view of `quorum-3` from the point of view of the other cluster nodes:

```
core@quorum-1 ~ $ for i in border quorum master worker; do loopssh ${i} arp -a | grep ^quorum-3; done
quorum-3.cell-1.dc-1.demo.lan (10.136.83.75) at 06:65:1f:74:8f:63 [ether] on eth0
quorum-3.cell-1.dc-1.demo.lan (10.136.83.75) at 06:65:1f:74:8f:63 [ether] on eth0
quorum-3.cell-1.dc-1.demo.lan (10.136.83.75) at 06:65:1f:74:8f:63 [ether] on eth0
quorum-3.cell-1.dc-1.demo.lan (10.136.83.75) at 06:65:1f:74:8f:63 [ether] on eth0
quorum-3.cell-1.dc-1.demo.lan (10.136.83.75) at 06:65:1f:74:8f:63 [ether] on eth0
quorum-3.cell-1.dc-1.demo.lan (10.136.83.75) at 06:65:1f:74:8f:63 [ether] on eth0
quorum-3.cell-1.dc-1.demo.lan (10.136.83.75) at 06:65:1f:74:8f:63 [ether] on eth0
quorum-3.cell-1.dc-1.demo.lan (10.136.83.75) at 06:65:1f:74:8f:63 [ether] on eth0
quorum-3.cell-1.dc-1.demo.lan (10.136.83.75) at 06:65:1f:74:8f:63 [ether] on eth0
```

## 2. Destroy the elected master

A few seconds after terminating the `quorum-3` *EC2* instance, `quorum-2` is elected for *zookeeper*:

```
core@quorum-1 ~ $ loopssh quorum "echo srvr | ncat \$(hostname) 2181 | grep Mode"
--[ quorum-1.cell-1.dc-1.demo.lan ]--
Mode: follower
--[ quorum-2.cell-1.dc-1.demo.lan ]--
Mode: leader
```

And `quorum-1` is elected for *etcd*:

```
core@quorum-1 ~ $ etcdctl member list
deff3e5ba168963: name=quorum-2 peerURLs=http://10.136.89.99:2380 clientURLs=http://10.136.89.99:2379 isLeader=false
498aec03aa857694: name=quorum-3 peerURLs=http://10.136.83.75:2380 clientURLs=http://10.136.83.75:2379 isLeader=false
fadca54d49882874: name=quorum-1 peerURLs=http://10.136.113.221:2380 clientURLs=http://10.136.113.221:2379 isLeader=true
```

I can still browse the *mesos* and *marathon* web GUIs. I can also see the expected information and the application is up, running and usable.

## 3. Purge the terminated instance:

Since there is no chance for this instance to come back again, it is important to purge the *ARP* cache:

```
core@quorum-1 ~ $ for i in border quorum master worker; do loopssh ${i} sudo arp -d quorum-3.cell-1.dc-1.demo.lan; done
core@quorum-1 ~ $ for i in border quorum master worker; do loopssh ${i} arp -a | grep ^quorum-3; done
quorum-3.cell-1.dc-1.demo.lan (10.136.83.75) at <incomplete> on eth0
quorum-3.cell-1.dc-1.demo.lan (10.136.83.75) at <incomplete> on eth0
quorum-3.cell-1.dc-1.demo.lan (10.136.83.75) at <incomplete> on eth0
quorum-3.cell-1.dc-1.demo.lan (10.136.83.75) at <incomplete> on eth0
quorum-3.cell-1.dc-1.demo.lan (10.136.83.75) at <incomplete> on eth0
quorum-3.cell-1.dc-1.demo.lan (10.136.83.75) at <incomplete> on eth0
quorum-3.cell-1.dc-1.demo.lan (10.136.83.75) at <incomplete> on eth0
quorum-3.cell-1.dc-1.demo.lan (10.136.83.75) at <incomplete> on eth0
quorum-3.cell-1.dc-1.demo.lan (10.136.83.75) at <incomplete> on eth0
```

Also I must notify the remaining members of the *etcd* cluster not to expect the lost node to come back again:

```
core@quorum-1 ~ $ etcdctl member remove 498aec03aa857694
Removed member 498aec03aa857694 from cluster
core@quorum-1 ~ $ etcdctl member list
deff3e5ba168963: name=quorum-2 peerURLs=http://10.136.89.99:2380 clientURLs=http://10.136.89.99:2379 isLeader=false
fadca54d49882874: name=quorum-1 peerURLs=http://10.136.113.221:2380 clientURLs=http://10.136.113.221:2379 isLeader=true
```

## 4. Create a brand new quorum node:

This is a new machine, it only shares the name with the previous one:

```
[0] ~ >> katoctl ec2 add --cluster-id cell-1-dub --host-id 3 --host-name quorum --instance-type m3.medium --roles quorum
INFO[0000] Latest CoreOS stable AMI located              cmd=ec2:add id=ami-b7cba3c4
INFO[0000] Rendering gzipped cloud-config template       cmd=udata id=quorum-3
INFO[0001] New m3.medium EC2 instance requested          cmd=ec2:run id=i-00c3028d
INFO[0001] New EC2 instance tagged
```

A few minutes later I can ssh into `quorum-3` again:

```
core@quorum-3 ~ $ katostat
LoadState=loaded  ActiveState=active  SubState=running  Id=etcd2.service
LoadState=loaded  ActiveState=active  SubState=running  Id=docker.service
LoadState=loaded  ActiveState=active  SubState=running  Id=zookeeper.service
LoadState=loaded  ActiveState=active  SubState=running  Id=rexray.service
LoadState=loaded  ActiveState=active  SubState=running  Id=cadvisor.service
LoadState=loaded  ActiveState=active  SubState=running  Id=node-exporter.service
LoadState=loaded  ActiveState=active  SubState=running  Id=zookeeper-exporter.service
LoadState=loaded  ActiveState=active  SubState=waiting  Id=etchost.timer
```

## 5. Add the new member to *etcd*:

Add the new member to the cluster:

```
core@quorum-1 ~ $ etcdctl member add quorum-3 http://10.136.117.252:2380
Added member named quorum-3 with ID ca42100b497845c0 to cluster

ETCD_NAME="quorum-3"
ETCD_INITIAL_CLUSTER="quorum-2=http://10.136.89.99:2380,quorum-3=http://10.136.117.252:2380,quorum-1=http://10.136.113.221:2380"
ETCD_INITIAL_CLUSTER_STATE="existing"
```

New member with new cluster configuration:

```
core@quorum-3 ~ $ cat /run/systemd/system/etcd2.service.d/20-cloudinit.conf
[Service]
Environment="ETCD_NAME=quorum-3"
Environment="ETCD_ADVERTISE_CLIENT_URLS=http://10.136.117.252:2379"
Environment="ETCD_LISTEN_CLIENT_URLS=http://127.0.0.1:2379,http://10.136.117.252:2379"
Environment="ETCD_LISTEN_PEER_URLS=http://10.136.117.252:2380"
Environment="ETCD_INITIAL_CLUSTER=quorum-2=http://10.136.89.99:2380,quorum-3=http://10.136.117.252:2380,quorum-1=http://10.136.113.221:2380"
Environment="ETCD_INITIAL_CLUSTER_STATE=existing"
```

Start the new member:

```
core@quorum-3 ~ $ sudo systemctl daemon-reload
core@quorum-3 ~ $ sudo systemctl start etcd2
```
