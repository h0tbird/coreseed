#### Deploy on Vagrant

```bash
export KATO_MASTER_COUNT=3
export KATO_NODE_COUNT=2
export KATO_EDGE_COUNT=0

vagrant up

ssh-add ~/.vagrant.d/insecure_private_key
TERM=xterm vagrant ssh master-1
```
