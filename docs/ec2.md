### Deploy on Amazon EC2

Before you start make sure:
- Your system's clock is synchronized.
- You have uploaded valid `SSH` keys to `EC2` ([doc](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html)).
- You have AWS credentials in `~/.aws/credentials` ([doc](https://github.com/aws/aws-sdk-go/wiki/configuring-sdk#shared-credentials-file)).
- You have permissions to manage `EC2` and `VPC` ([doc](http://docs.aws.amazon.com/IAM/latest/UserGuide/access_permissions.html)).

#### Environment
Define your environment. If you want to reuse existing EBS volumes you must target the EC2 region and availability zone where your volumes are stored:
```bash
export KATO_EC2_DEPLOY_CLUSTER_ID='<cluster-id>'
export KATO_EC2_DEPLOY_NS1_API_KEY='<ns1-private-key>'
export KATO_EC2_DEPLOY_DOMAIN='<ns1-managed-public-domain>'
export KATO_EC2_DEPLOY_REGION='<ec2-region>'
export KATO_EC2_DEPLOY_KEY_PAIR='<ec2-ssh-key-name>'
```

#### For operators
If you are an *operator* you need `the real thing`&trade;
```bash
katoctl ec2 deploy \
  --master-count 3 \
  --worker-count 2 \
  --edge-count 1 \
  --ns1-api-key ${KATO_EC2_DEPLOY_NS1_API_KEY} \
  --domain ${KATO_EC2_DEPLOY_DOMAIN} \
  --region ${KATO_EC2_DEPLOY_REGION} \
  --key-pair ${KATO_EC2_DEPLOY_KEY_PAIR}
```

#### For developers
If you are a *developer* you can deploy a lighter version:
```bash
katoctl ec2 deploy \
  --master-count 1 \
  --worker-count 1 \
  --edge-count 1 \
  --ns1-api-key ${KATO_EC2_DEPLOY_NS1_API_KEY} \
  --domain ${KATO_EC2_DEPLOY_DOMAIN} \
  --region ${KATO_EC2_DEPLOY_REGION} \
  --key-pair ${KATO_EC2_DEPLOY_KEY_PAIR}
```

#### Wait for it...
At this point you must wait for `EC2` to report healthy checks for all your instances. Now you're done deploying infrastructure, go back to step 3 in the main [README](https://github.com/h0tbird/kato/blob/master/README.md#3-pre-flight-checklist).
