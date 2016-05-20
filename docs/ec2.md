### Deploy on Amazon EC2

Before you start make sure:
- Your system's clock is synchronized.
- You have uploaded valid `SSH` keys to `EC2` ([doc](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html)).
- You have AWS credentials in `~/.aws/credentials` ([doc](https://github.com/aws/aws-sdk-go/wiki/configuring-sdk#shared-credentials-file)).
- You have permissions to manage `EC2` and `VPC` ([doc](http://docs.aws.amazon.com/IAM/latest/UserGuide/access_permissions.html)).

#### Environment
Define your environment. If you want to reuse existing EBS volumes you must target the EC2 region and availability zone where your volumes are stored:
```bash
export KATO_DEPLOY_EC2_NS1_API_KEY='<your-ns1-private-key>'
export KATO_DEPLOY_EC2_DOMAIN='<your-ns1-managed-public-domain>'
export KATO_DEPLOY_EC2_REGION='<your-ec2-region>'
export KATO_DEPLOY_EC2_ZONE='<your-ec2-availability-zone>'
export KATO_DEPLOY_EC2_KEY_PAIR='<your-ec2-ssh-key-name>'
export KATO_DEPLOY_EC2_COREOS_CHANNEL='<your-coreos-release-channel>'
```

#### For operators
If you are an *operator* you need `the real thing`&trade;
```bash
katoctl ec2 deploy \
  --master-count 3 \
  --node-count 2 \
  --edge-count 1 \
  --master-type t2.medium \
  --node-type m3.large \
  --edge-type t2.small \
  --ns1-api-key ${KATO_DEPLOY_EC2_NS1_API_KEY} \
  --domain ${KATO_DEPLOY_EC2_DOMAIN} \
  --region ${KATO_DEPLOY_EC2_REGION} \
  --zone ${KATO_DEPLOY_EC2_ZONE} \
  --key-pair ${KATO_DEPLOY_EC2_KEY_PAIR} \
  --channel ${KATO_DEPLOY_EC2_COREOS_CHANNEL}
```

#### For developers
If you are a *developer* you can deploy a lighter version:
```bash
katoctl ec2 deploy \
  --master-count 1 \
  --node-count 1 \
  --edge-count 1 \
  --node-type m3.large \
  --edge-type t2.small \
  --ns1-api-key ${KATO_DEPLOY_EC2_NS1_API_KEY} \
  --domain ${KATO_DEPLOY_EC2_DOMAIN} \
  --region ${KATO_DEPLOY_EC2_REGION} \
  --zone ${KATO_DEPLOY_EC2_ZONE} \
  --key-pair ${KATO_DEPLOY_EC2_KEY_PAIR} \
  --channel ${KATO_DEPLOY_EC2_COREOS_CHANNEL}
```

#### Wait for it...
At this point you must wait for `EC2` to report healthy checks for all your instances. Now you're done deploying infrastructure, go back to step 3 in the main [README](https://github.com/h0tbird/kato/blob/master/README.md#3-pre-flight-checklist).
