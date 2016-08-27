---
title: Installing Káto
---

# 1. Install katoctl

All you need to deploy *Káto* is the `katoctl` binary available in your *PATH*. If you are a end user just grab the latest stable release. If you are a *Káto* developer or you want to test the latest code you can get the go-getable version which follows the master branch (and might be broken). Also, you can optionally setup shell completion.

<div class="col-xs-12" style="height:10px;"></div>

<ul class="nav nav-tabs">
 <li class="active"><a href="#1" data-toggle="tab">Latest release</a></li>
 <li><a href="#2" data-toggle="tab">Main branch</a></li>
 <li><a href="#3" data-toggle="tab">Shell completion</a></li>
</ul>

<div class="tab-content ">
 <div class="tab-pane active" id="1">
  <div class="panel panel-default">
   <div class="panel-body">
    <center><em>curl kato.one/go | sh</em></center>
   </div>
  </div>
 </div>
 <div class="tab-pane" id="2">
  <div class="panel panel-default">
   <div class="panel-body">
    <center><em>go get -u github.com/katosys/kato/cmd/katoctl</em></center>
   </div>
  </div>
 </div>
 <div class="tab-pane" id="3">
  <div class="panel panel-default">
   <div class="panel-body">
    <center><em>eval "$(katoctl --completion-script-${0#-})"</em></center>
   </div>
  </div>
 </div>
</div>

<div class="col-xs-12" style="height:10px;"></div>

# 2. Deploy Káto's infrastructure

*Káto* can be deployed on a few *IaaS* providers. More providers are planned but feel free to send a pull request if your prefered provider is not supported yet. Find below deployment guides for each supported provider:

<div class="col-xs-12" style="height:10px;"></div>

<div class="btn-group btn-group-justified" role="group" aria-label="...">
  <div class="btn-group" role="group">
    <a class="btn btn-default" href="{{ site.baseurl}}/docs/vagrant.html"><font color="#428bca">Vagrant</font></a>
  </div>
  <div class="btn-group" role="group">
    <a class="btn btn-default" href="{{ site.baseurl}}/docs/packet.html"><font color="#428bca">Packet.net</font></a>
  </div>
  <div class="btn-group" role="group">
    <a class="btn btn-default" href="{{ site.baseurl}}/docs/ec2.html"><font color="#428bca">Amazon EC2</font></a>
  </div>
</div>

<div class="col-xs-12" style="height:30px;"></div>

# 3. Pre-flight checklist

Once you have deployed the infrastructure, run sanity checks to evaluate whether the cluster is ready for normal operation. Use the `border-1` node if you are in the cloud or the `kato-1` node if you are using *Vagrant*. Also find [here]({{ site.baseurl}}/docs/checklist.html) an extended check list if you need to troubleshoot the cluster.

```bash
your@desk-1 ~ $ ssh -A core@border-1.ext.<managed-public-domain>
core@border-1 ~ $ etcdctl cluster-health
core@border-1 ~ $ for i in border quorum master worker; do loopssh ${i} katostat; done
```

# 4. Setup pritunl

*Pritunl* is an *OpenVPN* server that provides secure access to *Káto*'s private networks.
Access your *Pritunl* WebGUI at `http://border-1.ext.<managed-public-domain>`
Make sure you setup udp port `18443` for VPN connections.
