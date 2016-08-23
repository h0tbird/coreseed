---
title: Installing Káto
---

# 1. Install katoctl

All you need to deploy *Káto* is the `katoctl` binary in your *PATH*. If you are a end user just grab the latest stable release. If you are a *Káto* developer or you want to test the latest code you can get the go-getable version which follows the master branch (and might be broken). Also, you can optionally setup shell completion.

#### From the latest release (for *Káto* end users)

```bash
curl -s https://raw.githubusercontent.com/katosys/kato/master/install | bash
```

#### From the source code (for *Káto* developers)

```bash
go get -u github.com/katosys/kato/cmd/katoctl
go install github.com/katosys/kato/cmd/katoctl
```

#### Setup bash/zsh shell completion

```bash
eval "$(katoctl --completion-script-${0#-})"
```

# 2. Deploy Káto's infrastructure

*Káto* can be deployed on a few *IaaS* providers. More providers are planned but feel free to send a pull request if your prefered provider is not supported yet. Find below deployment guides for each supported provider:

- [Vagrant]({{ site.baseurl}}/docs/vagrant.html)
- [Packet.net]({{ site.baseurl}}/docs/packet.html)
- [Amazon EC2]({{ site.baseurl}}/docs/ec2.html)

# 3. Pre-flight checklist

Once you have deployed the infrastructure, run sanity checks to evaluate whether the cluster is ready for normal operation. Use the `border-1` node if you are in the cloud or the `kato-1` node if you are using *Vagrant*. Also find [here](https://github.com/katosys/kato/blob/master/docs/checklist.md) an extended check list if you need to troubleshoot the cluster.

```bash
marc@desk-1 ~ $ ssh -A core@border-1.ext.<managed-public-domain>
core@border-1 ~ $ etcdctl cluster-health
core@border-1 ~ $ for i in border quorum master worker; do loopssh ${i} katostat; done
```

# 4. Setup pritunl

*Pritunl* is an *OpenVPN* server that provides secure access to *Káto*'s private networks.
Access your *Pritunl* WebGUI at `http://border-1.ext.<managed-public-domain>`
Make sure you setup udp port `18443` for VPN connections.
