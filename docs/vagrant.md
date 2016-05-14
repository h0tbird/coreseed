### Deploy on Vagrant

#### For operators
If you are an *operator* you need `the real thing`&trade;
```bash
export KATO_MASTER_COUNT=3
export KATO_NODE_COUNT=2
export KATO_EDGE_COUNT=1
```

#### For developers
If you are a *developer* you can deploy a lighter version:
```bash
export KATO_MASTER_COUNT=1
export KATO_NODE_COUNT=1
export KATO_EDGE_COUNT=0
```

#### Everyone
Export your *NS1* private API key and your managed public domain:
```bash
export KATO_NS1_API_KEY='<your-ns1-api-key>'
export KATO_DOMAIN='<your-ns1-managed-public-domain>'
```

Find below other options and its default values:
```bash
export KATO_MASTER_COUNT=1
export KATO_NODE_COUNT=1
export KATO_EDGE_COUNT=0
export KATO_MASTER_CPUS=1
export KATO_MASTER_MEMORY=1024
export KATO_NODE_CPUS=2
export KATO_NODE_MEMORY=2048
export KATO_EDGE_CPUS=1
export KATO_EDGE_MEMORY=512
export KATO_COREOS_CHANNEL=alpha
export KATO_COREOS_VERSION=current
export KATO_NS1_API_KEY=x
export KATO_DOMAIN=cell-1.dc-1.demo.lan
export KATO_CA_CERT=''
```

#### Start and connect
It is very convenient to add the private ssh key to the ssh agent forwarding before you ssh into the box:

```bash
vagrant up
ssh-add ~/.vagrant.d/insecure_private_key
vagrant ssh master-1
```

Congratulations, you have now deployed the infrastructure. Go back to step 3 in the main [README](https://github.com/h0tbird/kato/blob/master/README.md#3-pre-flight-checklist) and run the pre-flight checklist before you start the *Káto's* stack.

#### Manage /etc/hosts

This is optional but recommended. Edit your `/etc/hosts` so you don't have to wait for the public DNS to propagate. In *OSX* you can use [Gas Mask](http://clockwise.ee/):

![Gas Mask](https://dl.dropboxusercontent.com/u/29639331/kato/gasmask_vagrant.png)
