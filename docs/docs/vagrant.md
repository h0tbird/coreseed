---
title: Deploy on Vagrant
---

# Deploy on Vagrant

Currently *VirtualBox* is the only supported *Vagrant* provider. Running `vagrant up` will deploy an all-in-one version of *Káto*. Your host's `~/git/` directory will be mapped to the `/code/` directory inside the VM (for this operation to succeed you might be prompted for a password).

## Environment variables

Find below the defined environment variables and their default values:

```bash
KATO_CLUSTER_ID='cell-1-demo'
KATO_NODE_CPUS='2'
KATO_NODE_MEMORY='4096'
KATO_VERSION='v0.1.0-alpha'
KATO_COREOS_CHANNEL='stable'
KATO_COREOS_VERSION='current'
KATO_NS1_API_KEY='x'
KATO_DOMAIN='cell-1.dc-1.demo.lan'
KATO_CA_CERT=''
KATO_CODE_PATH='~/git/'
```

## Start and connect

It is as simple as bringing up the VM and ssh into it:

```bash
vagrant up
vagrant ssh kato-1
```

Congratulations! You have now deployed the infrastructure. Go back to step 3 in the [Install katoctl]({{ site.baseurl}}/docs) section and run the pre-flight checklist.

## Manage /etc/hosts

This is optional but recommended. Edit your `/etc/hosts` so you don't have to wait for the public *DNS* to propagate (if you are using *NS1*). In *OSX* you can use [Gas Mask](http://clockwise.ee/):

![Gas Mask](https://raw.githubusercontent.com/katosys/kato/master/imgs/gasmask.png)
