#### Deploy on Amazon EC2
Make sure your system's clock is synchronized:
```
timedatectl set-ntp true
```
If you are an *operator* you need the `real thing`&trade;:
```bash
NS1_API_KEY='ns1-private-key-goes-here'

katoctl deploy ec2 \
  --master-count 3 \
  --node-count 2 \
  --edge-count 1 \
  --channel alpha \
  --region eu-west-1 \
  --domain cell-1.dc-1.demo.com \
  --ns1-api-key ${NS1_API_KEY} \
  --etcd-token auto \
  --ca-cert certificates/cell-1.dc-1.demo.com/certs/server-crt.pem \
  --key-pair your-ec2-ssh-key-name
```

If you are a *developer* you need a lighter version:
```bash
NS1_API_KEY='ns1-private-key-goes-here'

katoctl deploy ec2 \
  --master-count 1 \
  --node-count 1 \
  --edge-count 1 \
  --channel alpha \
  --region eu-west-1 \
  --domain cell-1.dc-1.demo.com \
  --ns1-api-key ${NS1_API_KEY} \
  --etcd-token auto \
  --ca-cert certificates/cell-1.dc-1.demo.com/certs/server-crt.pem \
  --key-pair your-ec2-ssh-key-name
```
