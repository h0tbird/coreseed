### Deploy on Amazon EC2

Before you start make sure:
- Your system's clock is synchronized.
- You have uploaded valid `SSH` keys to `EC2` ([doc](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html)).
- You have AWS credentials in `~/.aws/credentials` ([doc](https://github.com/aws/aws-sdk-go/wiki/configuring-sdk#shared-credentials-file)).
- You have permissions to manage `EC2` and `VPC` ([doc](http://docs.aws.amazon.com/IAM/latest/UserGuide/access_permissions.html)).

#### The EC2 provider

*Káto* can be deployed on *EC2* via the `katoctl ec2` provider. The `deploy` subcommand will call other commands and subcommands such as `setup` and `add` in order to orchestrate the deployment. Find below the output of `katoctl ec2 --help`:

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
    Adds a new instance to a Káto cluster on EC2.

  ec2 run
    Starts a CoreOS instance on Amazon EC2.
```

#### Deploy

If you want to reuse existing *EBS* volumes you must target the `--region` and `--zone` where your volumes are stored:

```bash
katoctl ec2 deploy \
  --master-count 3 \
  --worker-count 2 \
  --cluster-id <unique-cluster-id> \
  --ns1-api-key <ns1-private-key> \
  --domain <ns1-managed-public-domain> \
  --region <ec2-region> \
  --key-pair <ec2-ssh-key-name>
```

#### Add more workers
Adding a third worker is as easy as running:
```bash
katoctl ec2 add \
  --cluster-id <unique-cluster-id> \
  --role worker \
  --host-id 3
```

#### Wait for it...
At this point you must wait for `EC2` to report healthy checks for all your instances. Now you're done deploying infrastructure, go back to step 3 in the main [README](https://github.com/h0tbird/kato/blob/master/README.md#3-pre-flight-checklist).
