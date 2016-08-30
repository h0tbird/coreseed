---
title: Deploy on Vagrant
---

# Deploy on Vagrant

Currently *VirtualBox* (on *OSX* and *Linux*) is the only supported *Vagrant* provider. Running `vagrant up` will deploy an all-in-one version of *Káto*. Your host's `~/git/` directory will be mapped to the `/code/` directory inside the *Káto* VM and `sudo` is used to edit your host's `/etc/hosts` file (for this operations to succeed you might be prompted for a password).

<div class="col-xs-12" style="height:10px;"></div>

<ul class="nav nav-tabs">
 <li class="active"><a href="#1" data-toggle="tab">1. Clone</a></li>
 <li><a href="#2" data-toggle="tab">2. Setup</a></li>
 <li><a href="#3" data-toggle="tab">3. Start</a></li>
</ul>

<div class="tab-content ">
 <div class="tab-pane active" id="1">
  <div class="panel panel-default">
   <div class="panel-body">
    <center><em>git clone https://github.com/katosys/kato.git && cd kato</em></center>
   </div>
  </div>
 </div>
 <div class="tab-pane" id="2">
  <div class="panel panel-default">
   <div class="panel-body language-bash highlighter-rouge">
   <p>If you want to override the default <code class="highlighter-rouge">Vagrantfile</code> settings you can export the following environment variablest (default values below):</p>
   <pre class="highlight"><code>
   <span class="nv">KATO_CLUSTER_ID</span><span class="o">=</span><span class="s1">'vagrant-kato'</span>  <span class="c"># Unique ID used to identify the cluster.</span>
   <span class="nv">KATO_NODE_CPUS</span><span class="o">=</span><span class="s1">'2'</span>              <span class="c"># Virtual CPU cores per cluster node.</span>
   <span class="nv">KATO_NODE_MEMORY</span><span class="o">=</span><span class="s1">'4096'</span>         <span class="c"># Megabytes of memory per cluster node.</span>
   <span class="nv">KATO_VERSION</span><span class="o">=</span><span class="s1">'v0.1.0-beta'</span>      <span class="c"># Version of katoctl to fetch.</span>
   <span class="nv">KATO_COREOS_CHANNEL</span><span class="o">=</span><span class="s1">'stable'</span>    <span class="c"># CoreOS release [stable | beta | alpha]</span>
   <span class="nv">KATO_COREOS_VERSION</span><span class="o">=</span><span class="s1">'current'</span>   <span class="c"># CoreOS release version [current | version]</span>
   <span class="nv">KATO_NS1_API_KEY</span><span class="o">=</span><span class="s1">'x'</span>            <span class="c"># NS1 private API key (optional).</span>
   <span class="nv">KATO_DOMAIN</span><span class="o">=</span><span class="s1">'cell-1.dc-1.kato'</span>  <span class="c"># Managed domain name.</span>
   <span class="nv">KATO_CA_CERT</span><span class="o">=</span><span class="s1">''</span>                 <span class="c"># Path to SSL certificate (optional).</span>
   <span class="nv">KATO_CODE_PATH</span><span class="o">=</span><span class="s1">'~/git/'</span>         <span class="c"># Path to host's code directory.</span>
   </code></pre>
   </div>
  </div>
 </div>
 <div class="tab-pane" id="3">
  <div class="panel panel-default">
   <div class="panel-body language-bash highlighter-rouge">
   <p>It is as simple as bringing up the <em>VM</em> and wait for the services to start (be patient):</p>
   <pre class="highlight"><code>vagrant up
vagrant ssh kato-1 -c <span class="s2">"watch katostat"</span></code></pre>
   </div>
  </div>
 </div>
</div>

<div class="col-xs-12" style="height:10px;"></div>

## What's next?

Congratulations! You have now deployed an all-in-one local *Káto* system. Use the links below to browse through your new local system:

<div class="col-xs-12" style="height:10px;"></div>

<div class="btn-group btn-group-justified" role="group" aria-label="...">
  <div class="btn-group" role="group">
    <a class="btn btn-default" href="http://master-1.cell-1.dc-1.kato:5050"><font color="#428bca">Mesos</font></a>
  </div>
  <div class="btn-group" role="group">
    <a class="btn btn-default" href="http://master-1.cell-1.dc-1.kato:8080"><font color="#428bca">Marathon</font></a>
  </div>
  <div class="btn-group" role="group">
    <a class="btn btn-default" href="http://master-1.cell-1.dc-1.kato:4194"><font color="#428bca">cAdvisor</font></a>
  </div>
  <div class="btn-group" role="group">
    <a class="btn btn-default" href="http://worker-1.cell-1.dc-1.kato:9090/haproxy?stats"><font color="#428bca">HAProxy</font></a>
  </div>
  <div class="btn-group" role="group">
    <a class="btn btn-default" href="http://master-1.cell-1.dc-1.kato:9191/targets"><font color="#428bca">Prometheus</font></a>
  </div>
</div>

<div class="col-xs-12" style="height:30px;"></div>

## Deploy a sample application

[The Voting App](https://github.com/katosys/the-voting-app) is a sample application. The original demo was designed to run on top of a *Swarm* scheduler with docker compose. This one has been modified to run on top of a *Mesos/Marathon* scheduler. Click on the links below to access the vote and results websites.

```bash
git clone --recursive https://github.com/katosys/the-voting-app.git
cd the-voting-app
export MARATHON_URL='http://master-1.cell-1.dc-1.kato:8080'
./bin/marathon start
sudo bash -c "echo 172.17.8.11 vote.thevotingapp.com >> /etc/hosts"
sudo bash -c "echo 172.17.8.11 results.thevotingapp.com >> /etc/hosts"
```

<div class="btn-group btn-group-justified" role="group" aria-label="...">
  <div class="btn-group" role="group">
    <a class="btn btn-default" href="http://vote.thevotingapp.com"><font color="#428bca">Vote</font></a>
  </div>
  <div class="btn-group" role="group">
    <a class="btn btn-default" href="http://results.thevotingapp.com"><font color="#428bca">Results</font></a>
  </div>
</div>
