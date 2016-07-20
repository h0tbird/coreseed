## Master destroy & restore

Let's destroy one master and recreate it from scratch. I have 1 `edge`, 3 `master` and 2 `worker` nodes up and running on `EC2`. I am also connected to the cluster via `pritunl` which is running on the `edge` node. The cluster is running `The Voting App` which is a 5 container demo application deployed with *Marathon*. This is the status of the cluster before I destroy the `master-1` node:

#### Pre-storm etcd status

*Etcd* is running in a non-error state:
```
core@master-1 ~ $ etcdctl cluster-health
member 84e935dcf8f8f34d is healthy: got healthy result from http://10.136.0.11:2379
member c9f43ad86b922c35 is healthy: got healthy result from http://10.136.0.12:2379
member cef07101ac391840 is healthy: got healthy result from http://10.136.0.13:2379
cluster is healthy
```

`master-1` is the *etcd* leader:
```
core@master-1 ~ $ etcdctl member list
84e935dcf8f8f34d: name=master-1 peerURLs=http://10.136.0.11:2380 clientURLs=http://10.136.0.11:2379 isLeader=true
c9f43ad86b922c35: name=master-2 peerURLs=http://10.136.0.12:2380 clientURLs=http://10.136.0.12:2379 isLeader=false
cef07101ac391840: name=master-3 peerURLs=http://10.136.0.13:2380 clientURLs=http://10.136.0.13:2379 isLeader=false
```

#### Pre-storm zookeeper status

*Zookeeper* is running in a non-error state:
```
core@master-1 ~ $ loopssh master "echo ruok | ncat \$(hostname) 2181; echo"
--[ master-1.cell-1.dub.xnood.com ]--
imok
--[ master-2.cell-1.dub.xnood.com ]--
imok
--[ master-3.cell-1.dub.xnood.com ]--
imok
```

`master-3` is the *zookeeper* leader:
```
core@master-1 ~ $ loopssh master "echo srvr | ncat \$(hostname) 2181 | grep Mode"
--[ master-1.cell-1.dub.xnood.com ]--
Mode: follower
--[ master-2.cell-1.dub.xnood.com ]--
Mode: follower
--[ master-3.cell-1.dub.xnood.com ]--
Mode: leader
```
