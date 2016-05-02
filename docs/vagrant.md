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
Export your `NS1` private API key and your managed public domain:
```bash
export KATO_NS1_API_KEY='<your-ns1-api-key>'
export KATO_DOMAIN='<your-ns1-managed-public-domain>'
```

Find below other options and its default values:
```bash
export KATO_MASTER_COUNT=3
export KATO_NODE_COUNT=2
export KATO_EDGE_COUNT=1
export KATO_MASTER_CPUS=2
export KATO_MASTER_MEMORY=1024
export KATO_NODE_CPUS=2
export KATO_NODE_MEMORY=1024
export KATO_EDGE_CPUS=2
export KATO_EDGE_MEMORY=1024
export KATO_COREOS_CHANNEL=alpha
export KATO_COREOS_VERSION=current
export KATO_NS1_API_KEY=aabbccddeeaabbccddee
export KATO_DOMAIN=cell-1.dc-1.demo.lan
export KATO_CA_CERT=''
```

#### Start Vagrant
The REX-Ray driver leverages the `vboxwebserv` HTTP SOAP API which is a process that must be started from the VirtualBox host. It is optional to leverage authentication and it can be disabled:

```bash
VBoxManage setproperty websrvauthlibrary null
vboxwebsrv -H 0.0.0.0 -b
vagrant up
```

#### Connect
It is very convenient to add the private ssh key to the ssh agent forwarding before you ssh into the box:

```bash
ssh-add ~/.vagrant.d/insecure_private_key
vagrant ssh master-1
```

Congratulations, you have now deployed the infrastructure. Go back to step 3 in the main [README](https://github.com/h0tbird/kato/blob/master/README.md) and run the pre-flight checklist before you start the *Káto's* stack.

#### Gas Mask

This is optional but helpful. Manage your `/etc/hosts` with [Gas Mask](http://clockwise.ee/) so you don't have to wait for the public DNS to propagate.

![Gas Mask](https://dl.dropboxusercontent.com/u/29639331/kato/gasmask_vagrant.png)
