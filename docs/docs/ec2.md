---
title: Deploy on EC2
---

# Deploy on Amazon EC2

Before you start make sure:

- Your system's clock is synchronized.
- You have uploaded valid `SSH` keys to `EC2` ([doc](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html#how-to-generate-your-own-key-and-import-it-to-aws)).
- You have *AWS* credentials in `~/.aws/credentials` ([doc](https://docs.aws.amazon.com/sdk-for-go/v1/developerguide/configuring-sdk.html)).
- You have permissions to manage `IAM`, `VPC` and `EC2` ([doc](http://docs.aws.amazon.com/IAM/latest/UserGuide/access_permissions.html)).

## The EC2 provider

*Káto* can be deployed on *EC2* via the `katoctl ec2` provider. The `deploy` subcommand will recursively call other `katoctl` subcommands such as `ec2 setup` and `ec2 add` in order to orchestrate the deployment. Find below the output of `katoctl ec2 --help` for your reference:

```
usage: katoctl ec2 <command> [<args> ...]

This is the Káto EC2 provider.

Flags:
  -h, --help     Show context-sensitive help (also try --help-long and --help-man).
      --version  Show application version.

Subcommands:
  ec2 deploy
    Deploy Káto's infrastructure on Amazon EC2.

  ec2 setup
    Setup IAM, VPC and EC2 components.

  ec2 add
    Adds a new instance to an existing Káto cluster on EC2.

  ec2 run
    Starts a CoreOS instance on Amazon EC2.
```

## Deploy

If you want to reuse existing *EBS* volumes you must target the `--region` and `--zone` where your volumes are stored. During the deployment a cluster state file will be generated in your home directory under `~/.kato/<cluster-id>.json`.

<ul class="nav nav-tabs">
 <li class="active"><a href="#1" data-toggle="tab">Simple deploy example</a></li>
 <li><a href="#2" data-toggle="tab">Advanced deploy example</a></li>
 <li><a href="#3" data-toggle="tab">Add more nodes</a></li>
</ul>

<div class="tab-content ">
 <div class="tab-pane active" id="1">
  <div class="panel panel-default">
   <div class="panel-body language-bash highlighter-rouge">
    <p>Deploy one <code class="highlighter-rouge">m3.xlarge</code> <em>EC2</em> instance named <code class="highlighter-rouge">kato-1</code> with 3 <em>Káto</em> roles: <code class="highlighter-rouge">quorum</code>, <code class="highlighter-rouge">master</code> and <code class="highlighter-rouge">worker</code>. Also deploy a second <code class="highlighter-rouge">m3.medium</code> instance named <code class="highlighter-rouge">border-1</code> with the <code class="highlighter-rouge">border</code> <em>Káto</em> role assigned to it:</p>
    <pre class="highlight"><code>
 katoctl ec2 deploy
   --cluster-id &lt;cluster-id&gt; <span class="se">\</span>
   --ns1-api-key &lt;ns1-private-key&gt; <span class="se">\</span>
   --domain &lt;managed-public-domain&gt; <span class="se">\</span>
   --region &lt;ec2-region&gt; <span class="se">\</span>
   --key-pair &lt;ec2-ssh-key-name&gt; <span class="se">\</span>
   1:m3.xlarge:kato:quorum,master,worker <span class="se">\</span>
   1:m3.medium:border:border
    </code></pre>
   </div>
  </div>
 </div>

 <div class="tab-pane" id="2">
  <div class="panel panel-default">
   <div class="panel-body language-bash highlighter-rouge">
    <p>Find below a much more complex deploy where many options are set:</p>
    <pre class="highlight"><code>
 <span class="nb">export </span><span class="nv">KATO_EC2_DEPLOY_VPC_CIDR_BLOCK</span><span class="o">=</span><span class="s1">'10.136.0.0/16'</span>
 <span class="nb">export </span><span class="nv">KATO_EC2_DEPLOY_INTERNAL_SUBNET_CIDR</span><span class="o">=</span><span class="s1">'10.136.0.0/18'</span>
 <span class="nb">export </span><span class="nv">KATO_EC2_DEPLOY_EXTERNAL_SUBNET_CIDR</span><span class="o">=</span><span class="s1">'10.136.64.0/18'</span>
 <span class="nb">export </span><span class="nv">KATO_EC2_DEPLOY_FLANNEL_NETWORK</span><span class="o">=</span><span class="s1">'10.136.128.0/18'</span>
 <span class="nb">export </span><span class="nv">KATO_EC2_DEPLOY_FLANNEL_SUBNET_MIN</span><span class="o">=</span><span class="s1">'10.136.128.0'</span>
 <span class="nb">export </span><span class="nv">KATO_EC2_DEPLOY_FLANNEL_SUBNET_MAX</span><span class="o">=</span><span class="s1">'10.136.191.192'</span>
 <span class="nb">export </span><span class="nv">KATO_EC2_DEPLOY_FLANNEL_SUBNET_LEN</span><span class="o">=</span><span class="s1">'26'</span>

 katoctl ec2 deploy <span class="se">\</span>
   --cluster-id &lt;cluster-id&gt; <span class="se">\</span>
   --admin-email &lt;notifications-email&gt; <span class="se">\</span>
   --smtp-url &lt;smtp://user:pass@host:port&gt; <span class="se">\</span>
   --ns1-api-key &lt;ns1-private-key&gt; <span class="se">\</span>
   --sysdig-access-key &lt;sysdig-access-key&gt; <span class="se">\</span>
   --datadog-api-key &lt;datadog-api-key&gt; <span class="se">\</span>
   --slack-webhook &lt;slack-webhook-url&gt; <span class="se">\</span>
   --ca-cert &lt;path-to-crt-pem&gt; <span class="se">\</span>
   --stub-zone foo.demo.lan/192.168.1.201:53,192.168.1.202:53 <span class="se">\</span>
   --stub-zone bar.demo.lan/192.168.2.201:53,192.168.2.202:53 <span class="se">\</span>
   --domain &lt;managed-public-domain&gt; <span class="se">\</span>
   --region &lt;ec2-region&gt; <span class="se">\</span>
   --key-pair &lt;ec2-ssh-key-name&gt; <span class="se">\</span>
   3:m3.medium:quorum:quorum <span class="se">\</span>
   3:m3.medium:master:master <span class="se">\</span>
   3:m3.large:worker:worker <span class="se">\</span>
   1:m3.medium:border:border
    </code></pre>
   </div>
  </div>
 </div>

 <div class="tab-pane" id="3">
  <div class="panel panel-default">
   <div class="panel-body language-bash highlighter-rouge">
    <p>The cluster state file is read by <code class="highlighter-rouge">katoctl ec2 add</code>, adding a third worker is as easy as running:</p>
    <pre class="highlight"><code>
 katoctl ec2 add <span class="se">\</span>
   --cluster-id &lt;cluster-id&gt; <span class="se">\</span>
   --host-name worker <span class="se">\</span>
   --host-id 3 <span class="se">\</span>
   --roles worker <span class="se">\</span>
   --instance-type m3.large
    </code></pre>
   </div>
  </div>
 </div>

</div>

## Wait for it...
At this point you must wait for `EC2` to report healthy checks for all your instances. Now you're done deploying infrastructure, go back to step 3 in the [Install katoctl]({{ site.baseurl}}/docs) section.
