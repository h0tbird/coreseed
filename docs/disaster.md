## Elected master destroy & restore

Let's destroy the elected master and recreate it from scratch. I have 1 `border`, 3 `quorum`, 3 `master` and 3 `worker` nodes up and running on `EC2`. I am also connected to the cluster via `pritunl` which is running on the `border` node. The cluster is running `The Voting App` which is a 5 container demo application deployed with *Marathon*.

#### 1. Who is the elected master?

The answer is `master-1` for mesos:
```
core@master-1 ~ $ for i in 1 2 3; do curl -sI http://master-${i}:5050/redirect | grep Location; done
Location: //master-1.cell-1.dc-1.demo.lan:5050
Location: //master-1.cell-1.dc-1.demo.lan:5050
Location: //master-1.cell-1.dc-1.demo.lan:5050
```

And conveniently `master-1` for marathon too:
```
core@master-1 ~ $ for i in 1 2 3; do curl -s "http://master-${i}:8080/v2/leader" | jq '.'; done
{
  "leader": "master-1.cell-1.dc-1.demo.lan:8080"
}
{
  "leader": "master-1.cell-1.dc-1.demo.lan:8080"
}
{
  "leader": "master-1.cell-1.dc-1.demo.lan:8080"
}
```

This is the ARP view of `master-1` from the other cluster nodes:
```
core@master-1 ~ $ for i in border quorum master worker; do loopssh ${i} arp -a | grep ^master-1; done
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at 06:5f:d8:1e:5f:a9 [ether] on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at 06:5f:d8:1e:5f:a9 [ether] on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at 06:5f:d8:1e:5f:a9 [ether] on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at 06:5f:d8:1e:5f:a9 [ether] on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at 06:5f:d8:1e:5f:a9 [ether] on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at 06:5f:d8:1e:5f:a9 [ether] on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at 06:5f:d8:1e:5f:a9 [ether] on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at 06:5f:d8:1e:5f:a9 [ether] on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at 06:5f:d8:1e:5f:a9 [ether] on eth0
```

#### 2. Destroy the elected master

A few seconds after terminating the `master-1` `EC2` instance, `master-3` is elected for mesos:
```
core@master-2 ~ $ for i in 2 3; do curl -sI http://master-${i}:5050/redirect | grep Location; done
Location: //master-3.cell-1.dc-1.demo.lan:5050
Location: //master-3.cell-1.dc-1.demo.lan:5050
```

And also for marathon:
```
core@master-2 ~ $ for i in 2 3; do curl -s "http://master-${i}:8080/v2/leader" | jq '.'; done
{
  "leader": "master-3.cell-1.dc-1.demo.lan:8080"
}
{
  "leader": "master-3.cell-1.dc-1.demo.lan:8080"
}
```

I can still browse the mesos and marathon web GUIs. I can also see the expected information and the application is up, running and usable.

#### 3. Purge the terminated instance:

Since there is no chance for this instance to come back again, it is important to purge the ARP cache:
```
core@master-2 ~ $ for i in border quorum master worker; do loopssh ${i} sudo arp -d master-1.cell-1.dc-1.demo.lan; done
core@master-2 ~ $ for i in border quorum master worker; do loopssh ${i} arp -a | grep ^master-1; done
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at <incomplete> on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at <incomplete> on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at <incomplete> on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at <incomplete> on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at <incomplete> on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at <incomplete> on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at <incomplete> on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at <incomplete> on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at <incomplete> on eth0
```

#### 4. Create a brand new `master-1`:

This is a new machine, it only shares the name and IP with the previous one:
```
[0] ~ >> katoctl ec2 add --cluster-id cell-1-dub --host-id 1 --host-name master --instance-type m3.medium --roles master
INFO[0000] Latest CoreOS stable AMI located              cmd=ec2:add id=ami-b7cba3c4
INFO[0000] Rendering gzipped cloud-config template       cmd=udata id=master-1
INFO[0000] New m3.medium EC2 instance requested          cmd=ec2:run id=i-52fd3bdf
INFO[0001] New EC2 instance tagged
```

A few minutes later I can ssh into `master-1` again:
```
core@master-1 ~ $ katostat
LoadState=loaded  ActiveState=active  SubState=running  Id=etcd2.service
LoadState=loaded  ActiveState=active  SubState=running  Id=flanneld.service
LoadState=loaded  ActiveState=active  SubState=running  Id=docker.service
LoadState=loaded  ActiveState=active  SubState=running  Id=rexray.service
LoadState=loaded  ActiveState=active  SubState=running  Id=mesos-master.service
LoadState=loaded  ActiveState=active  SubState=running  Id=mesos-dns.service
LoadState=loaded  ActiveState=active  SubState=running  Id=marathon.service
LoadState=loaded  ActiveState=active  SubState=running  Id=confd.service
LoadState=loaded  ActiveState=active  SubState=running  Id=prometheus.service
LoadState=loaded  ActiveState=active  SubState=running  Id=cadvisor.service
LoadState=loaded  ActiveState=active  SubState=running  Id=mesos-master-exporter.service
LoadState=loaded  ActiveState=active  SubState=running  Id=node-exporter.service
LoadState=loaded  ActiveState=active  SubState=waiting  Id=etchost.timer
```

This is the new ARP view:
```
core@master-1 ~ $ for i in border quorum master worker; do loopssh ${i} arp -a | grep ^master-1; done
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at 06:66:72:e3:21:77 [ether] on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at 06:66:72:e3:21:77 [ether] on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at 06:66:72:e3:21:77 [ether] on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at 06:66:72:e3:21:77 [ether] on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at 06:66:72:e3:21:77 [ether] on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at 06:66:72:e3:21:77 [ether] on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at 06:66:72:e3:21:77 [ether] on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at 06:66:72:e3:21:77 [ether] on eth0
master-1.cell-1.dc-1.demo.lan (10.136.64.11) at 06:66:72:e3:21:77 [ether] on eth0
```

And the elected master still is `master-3`:
```
core@master-1 ~ $ for i in 1 2 3; do curl -sI http://master-${i}:5050/redirect | grep Location; done
Location: //master-3.cell-1.dc-1.demo.lan:5050
Location: //master-3.cell-1.dc-1.demo.lan:5050
Location: //master-3.cell-1.dc-1.demo.lan:5050
```

#### 5. Force `master-1` to become the elected master again:

Stop `mesos-master` and `marathon` on `master-3`:
```
core@master-3 ~ $ sudo systemctl stop mesos-master; sleep 10; sudo systemctl start mesos-master
core@master-3 ~ $ sudo systemctl stop marathon; sleep 10; sudo systemctl start marathon
```

The new mesos elected master is `master-2`:
```
core@master-1 ~ $ for i in 1 2 3; do curl -sI http://master-${i}:5050/redirect | grep Location; done
Location: //master-2.cell-1.dc-1.demo.lan:5050
Location: //master-2.cell-1.dc-1.demo.lan:5050
Location: //master-2.cell-1.dc-1.demo.lan:5050
```

The new marathon elected master is `master-1`:
```
core@master-1 ~ $ for i in 1 2 3; do curl -s "http://master-${i}:8080/v2/leader" | jq '.'; done
{
  "leader": "master-1.cell-1.dc-1.demo.lan:8080"
}
{
  "leader": "master-1.cell-1.dc-1.demo.lan:8080"
}
{
  "leader": "master-1.cell-1.dc-1.demo.lan:8080"
}
```

Stop `mesos-master` on `master-2`:
```
core@master-2 ~ $ sudo systemctl stop mesos-master; sleep 10; sudo systemctl start mesos-master
```

The new mesos elected master is `master-1` and everything works as expected:
```
core@master-1 ~ $ for i in 1 2 3; do curl -sI http://master-${i}:5050/redirect | grep Location; done
Location: //master-1.cell-1.dc-1.demo.lan:5050
Location: //master-1.cell-1.dc-1.demo.lan:5050
Location: //master-1.cell-1.dc-1.demo.lan:5050
```
