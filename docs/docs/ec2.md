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

If you want to reuse existing *EBS* volumes you must target the `--region` and `--zone` where your volumes are stored. During the deployment a state file will be generated under `$HOME/.kato/<unique-cluster-id>.json`

### Simple deploy example:
```bash
katoctl ec2 deploy \
  --cluster-id <unique-cluster-id> \
  --ns1-api-key <ns1-private-key> \
  --domain <managed-public-domain> \
  --region <ec2-region> \
  --key-pair <ec2-ssh-key-name> \
  1:m3.xlarge:kato:quorum,master,worker \
  1:m3.medium:border:border
```

### Complex deploy example:
```bash
export KATO_EC2_DEPLOY_VPC_CIDR_BLOCK='10.136.0.0/16'
export KATO_EC2_DEPLOY_INTERNAL_SUBNET_CIDR='10.136.0.0/18'
export KATO_EC2_DEPLOY_EXTERNAL_SUBNET_CIDR='10.136.64.0/18'
export KATO_EC2_DEPLOY_FLANNEL_NETWORK='10.136.128.0/18'
export KATO_EC2_DEPLOY_FLANNEL_SUBNET_MIN='10.136.128.0'
export KATO_EC2_DEPLOY_FLANNEL_SUBNET_MAX='10.136.191.192'
export KATO_EC2_DEPLOY_FLANNEL_SUBNET_LEN='26'

katoctl ec2 deploy \
  --cluster-id <unique-cluster-id> \
  --ns1-api-key <ns1-private-key> \
  --sysdig-access-key <sysdig-access-key> \
  --datadog-api-key <datadog-api-key> \
  --ca-cert <path-to-crt-pem> \
  --stub-zone foo.demo.lan/192.168.1.201:53,192.168.1.202:53 \
  --stub-zone bar.demo.lan/192.168.2.201:53,192.168.2.202:53 \
  --domain <managed-public-domain> \
  --region <ec2-region> \
  --key-pair <ec2-ssh-key-name> \
  3:m3.medium:quorum:quorum \
  3:m3.medium:master:master \
  3:m3.large:worker:worker \
  1:m3.medium:border:border
```

## Add more workers

The state file is read by `katoctl ec2 add`, adding a third worker is as easy as running:

```bash
katoctl ec2 add \
  --cluster-id <unique-cluster-id> \
  --host-name worker \
  --host-id 3 \
  --roles worker \
  --instance-type m3.large
```

## Wait for it...
At this point you must wait for `EC2` to report healthy checks for all your instances. Now you're done deploying infrastructure, go back to step 3 in the [Install katoctl]({{ site.baseurl}}/docs) section.
