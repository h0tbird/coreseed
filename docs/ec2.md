### Deploy on Amazon EC2

Before you start make sure:
- Your system's clock is synchronized.
- You have uploaded valid `SSH` keys to `EC2`.
- You have AWS credentials in `~/.aws/credentials`.
- You have permissions to manage `EC2` and `VPC`.

Define your environment:
```bash
KATO_DEPLOY_EC2_NS1_API_KEY='<your-ns1-private-key-goes-here>'
KATO_DEPLOY_EC2_DOMAIN='<your-public-domain-goes-here>'
KATO_DEPLOY_EC2_KEY_PAIR='<your-ec2-ssh-key-name-goes-here>'
```

#### For operators
If you are an *operator* you need `the real thing`&trade;
```bash
katoctl deploy ec2 \
  --master-count 3 \
  --node-count 2 \
  --edge-count 1 \
  --channel alpha \
  --region eu-west-1 \
  --domain ${KATO_DEPLOY_EC2_DOMAIN} \
  --ns1-api-key ${KATO_DEPLOY_EC2_NS1_API_KEY} \
  --key-pair ${KATO_DEPLOY_EC2_KEY_PAIR}
```

#### For developers
If you are a *developer* you can deploy a lighter version:
```bash
katoctl deploy ec2 \
  --master-count 1 \
  --node-count 1 \
  --edge-count 1 \
  --channel alpha \
  --region eu-west-1 \
  --domain ${KATO_DEPLOY_EC2_DOMAIN} \
  --ns1-api-key ${KATO_DEPLOY_EC2_NS1_API_KEY} \
  --key-pair ${KATO_DEPLOY_EC2_KEY_PAIR}
```
