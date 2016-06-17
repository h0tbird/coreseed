## Extended checklist

Find below a list of actions used to troubleshoot *Káto*, this list is based on real issues and their solutions.
Some issues might have been permanently fixed but are keept here for its troubleshooting value.

###### **PROBLEM:** *The resource demand for a given task is higher than the available resources co-located on a single worker node. Therefore, the Marathon task stays in the waiting state forever.*

This is not really an error, you can:
- Exo-scale up your cluster.
- Redefine the task so it requires less resources.
- Kill existing tasks in order to free resources.

###### **PROBLEM:** *Multiple Marathon frameworks registered but only one is expected to be up and running.*

Try to teardown the unexpected framework ID:

```
curl -sX POST http://master.mesos:5050/master/teardown -d 'frameworkId=aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee-ffff'
```

###### **PROBLEM:** *Containers on one worker are unable to ping containers on other workers.*

This is most likely to be a *docker*-*fleet* communication problem. Was *fleet* up and running at the time *docker* started?
Run this command to check the IP address assigned by *fleet* to the *docker0* bridge:

```
loopssh ip r | grep docker0
```

Restart *docker* if the IP address is not within the range managed by *fleet*.
