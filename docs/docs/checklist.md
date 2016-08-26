---
title: Troubleshooting
---

# Troubleshooting

Find below a list of actions used to troubleshoot *Káto*, this list is based on real issues and their solutions.
Some issues might have been permanently fixed but are keept here for its troubleshooting value.

<br>
<h4><span class="glyphicon glyphicon glyphicon-pencil" aria-hidden="true"></span> <em>Multiple HA-Proxy PIDs inside one container</em></h4>
<hr>

**Diagnose:**

```
loopssh worker "docker exec -i marathon-lb ps auxf | grep 'haproxy -p'"
```

**Mitigate:**

```
```

<br>
<h4><span class="glyphicon glyphicon glyphicon-pencil" aria-hidden="true"></span> <em>Disk usage</em></h4>
<hr>

**Diagnose:**
```
for i in quorum master worker border; do loopssh ${i} df -h; done
```

**Mitigate:**
```
sudo journalctl --vacuum-time=1h
docker rmi $(docker images -qf dangling=true)
```

<br>
<h4><span class="glyphicon glyphicon glyphicon-pencil" aria-hidden="true"></span> <em>Mixed CoreOS versions</em></h4>
<hr>

**Diagnose:**
```
for i in quorum master worker border; do
  loopssh ${i} "cat /etc/os-release | grep VERSION="
done
```

**Mitigate:**
```
update_engine_client -check_for_update
```

<br>
<h4><span class="glyphicon glyphicon glyphicon-pencil" aria-hidden="true"></span> <em>Summary of running containers (not realtime)</em></h4>
<hr>

```
for i in $(etcdctl ls /docker/images); do etcdctl get ${i}; done | \
sort | uniq -c | sort -n
```

<br>
<h4><span class="glyphicon glyphicon glyphicon-pencil" aria-hidden="true"></span> <em>The resource demand for a given task is higher than the available resources co-located on a single worker node. Therefore, the Marathon task stays in the waiting state forever.</em></h4>
<hr>

This is not really an error, you can:
- Exo-scale up your cluster.
- Redefine the task so it requires less resources.
- Kill existing tasks in order to free resources.

<br>
<h4><span class="glyphicon glyphicon glyphicon-pencil" aria-hidden="true"></span> <em>Multiple Marathon frameworks registered but only one is expected to be up and running.</em></h4>
<hr>

Try to teardown the unexpected framework ID:

```
curl \
  -sX POST http://master.mesos:5050/master/teardown \
  -d 'frameworkId=aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee-ffff'
```

<br>
<h4><span class="glyphicon glyphicon glyphicon-pencil" aria-hidden="true"></span> <em>Containers on one worker are unable to ping containers on the other workers.</em></h4>
<hr>

This is most likely to be a *docker*-*fleet* communication problem. Was *fleet* up and running at the time *docker* started? Run the command below to check whether the IP address assigned by *fleet* to the *docker0* bridge is within the range managed by *fleet*, restart *docker* otherwise:

```
loopssh worker "ip r | grep docker0"
```
