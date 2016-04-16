#### Deploy on Vagrant

If you are an *operator* you need the `real thing`&trade;:
```bash
export KATO_MASTER_COUNT=3
export KATO_NODE_COUNT=2
export KATO_EDGE_COUNT=1
```

If you are a *developer* you need a lighter version:
```bash
export KATO_MASTER_COUNT=1
export KATO_NODE_COUNT=1
export KATO_EDGE_COUNT=0
```

Start and connect:
```bash
vagrant up
ssh-add ~/.vagrant.d/insecure_private_key
TERM=xterm vagrant ssh master-1
```
