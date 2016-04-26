### Deploy on Amazon EC2

Before you start make sure:
- Your system's clock is synchronized.
- You have uploaded valid `SSH` keys to `EC2`.
- You have AWS credentials in `~/.aws/credentials`.
- You have permissions to manage `EC2` and `VPC`.

#### Environment
Define your environment:
```bash
export KATO_DEPLOY_EC2_NS1_API_KEY='<your-ns1-private-key-goes-here>'
export KATO_DEPLOY_EC2_DOMAIN='<your-public-domain-goes-here>'
export KATO_DEPLOY_EC2_REGION='<your-ec2-region-goes-here>'
export KATO_DEPLOY_EC2_KEY_PAIR='<your-ec2-ssh-key-name-goes-here>'
export KATO_DEPLOY_EC2_CHANNEL='<your-coreos-release-channel>'
```

#### For operators
If you are an *operator* you need `the real thing`&trade;
```bash
katoctl deploy ec2 \
  --master-count 3 \
  --node-count 2 \
  --edge-count 1 \
  --master-type t2.medium \
  --node-type m3.large \
  --edge-type t2.small \
  --ns1-api-key ${KATO_DEPLOY_EC2_NS1_API_KEY} \
  --domain ${KATO_DEPLOY_EC2_DOMAIN} \
  --region ${KATO_DEPLOY_EC2_REGION} \
  --key-pair ${KATO_DEPLOY_EC2_KEY_PAIR} \
  --channel ${KATO_DEPLOY_EC2_CHANNEL}
```

#### For developers
If you are a *developer* you can deploy a lighter version:
```bash
katoctl deploy ec2 \
  --master-count 1 \
  --node-count 1 \
  --edge-count 1 \
  --node-type m3.large \
  --edge-type t2.small \
  --ns1-api-key ${KATO_DEPLOY_EC2_NS1_API_KEY} \
  --domain ${KATO_DEPLOY_EC2_DOMAIN} \
  --region ${KATO_DEPLOY_EC2_REGION} \
  --key-pair ${KATO_DEPLOY_EC2_KEY_PAIR} \
  --channel ${KATO_DEPLOY_EC2_CHANNEL}
```
