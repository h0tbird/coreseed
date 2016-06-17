## Extended checklist

Find below a list of actions used to troubleshoot *Káto*, this list is based on real issues and their solutions.
Some issues might have been permanently fixed but are keept here for its troubleshooting value.

##### PROBLEM: The resource demand for a given task is higher than the available resources co-located on a single worker node. Therefore, the Marathon task stays in the waiting state forever.

This is not really an error, you can:
- Exo-scale up your cluster.
- Redefine the task so it requires less resources.
- Kill existing tasks in order to free resources.

##### PROBLEM: Multiple Marathon frameworks registered but only one is expected and running.

Try to teardown the unexpected framework ID:

```
curl -sX POST http://master.mesos:5050/master/teardown -d 'frameworkId=aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee-ffff'
```

##### PROBLEM:
