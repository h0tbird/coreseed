### Deploy on Amazon EC2

Before you start make sure:
- Your system's clock is synchronized.
- You have uploaded valid SSH keys to EC2.
- You have AWS credentials in `~/.aws/credentials`.
- You have permissions to manage `EC2` and `VPC`.

Define your environment:
```bash
NS1_API_KEY='<ns1-private-key-goes-here>'
DOMAIN='<your-public-domain-goes-here>'
KEY_PAIR='<your-ec2-ssh-key-name-goes-here>'
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
  --domain ${DOMAIN} \
  --ns1-api-key ${NS1_API_KEY} \
  --etcd-token auto \
  --key-pair ${KEY_PAIR}
```

#### For developers
If you are a *developer* you need a lighter version:
```bash
katoctl deploy ec2 \
  --master-count 1 \
  --node-count 1 \
  --edge-count 1 \
  --channel alpha \
  --region eu-west-1 \
  --domain ${DOMAIN} \
  --ns1-api-key ${NS1_API_KEY} \
  --etcd-token auto \
  --key-pair ${KEY_PAIR}
```
