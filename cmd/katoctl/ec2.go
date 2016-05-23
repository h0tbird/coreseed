package main

//-----------------------------------------------------------------------------
// 'katoctl ec2' command flags definitions:
//-----------------------------------------------------------------------------

var (

	//------------------------
	// ec2: top level command
	//------------------------

	cmdEc2 = app.Command("ec2", "Kato's EC2 provider.")

	//----------------------------
	// ec2 deploy: nested command
	//----------------------------

	cmdEc2Deploy = cmdEc2.Command("deploy", "Deploy Kato's infrastructure on Amazon EC2.")

	flEc2DeployClusterID = cmdEc2Deploy.Flag("cluster-id", "Cluster ID for later reference.").
				Required().PlaceHolder("KATO_DEPLOY_EC2_CLUSTER_ID").
				OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_CLUSTER_ID").
				String()

	flEc2DeployMasterCount = cmdEc2Deploy.Flag("master-count", "Number of master nodes to deploy [ 1 | 3 | 5 ]").
				Required().PlaceHolder("KATO_DEPLOY_EC2_MASTER_COUNT").
				OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_MASTER_COUNT").
				Short('m').HintOptions("1", "3", "5").Int()

	flEc2DeployNodeCount = cmdEc2Deploy.Flag("node-count", "Number of worker nodes to deploy.").
				Required().PlaceHolder("KATO_DEPLOY_EC2_NODE_COUNT").
				OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_NODE_COUNT").
				Short('n').Int()

	flEc2DeployEdgeCount = cmdEc2Deploy.Flag("edge-count", "Number of edge nodes to deploy.").
				Required().PlaceHolder("KATO_DEPLOY_EC2_EDGE_COUNT").
				OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_EDGE_COUNT").
				Short('e').Int()

	flEc2DeployMasterType = cmdEc2Deploy.Flag("master-type", "EC2 master instance type.").
				Default("t2.medium").OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_MASTER_TYPE").
				String()

	flEc2DeployNodeType = cmdEc2Deploy.Flag("node-type", "EC2 node instance type.").
				Default("t2.medium").OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_NODE_TYPE").
				String()

	flEc2DeployEdgeType = cmdEc2Deploy.Flag("edge-type", "EC2 edge instance type.").
				Default("t2.medium").OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_EDGE_TYPE").
				String()

	flEc2DeployChannel = cmdEc2Deploy.Flag("channel", "CoreOS release channel [ stable | beta | alpha ]").
				Required().PlaceHolder("KATO_DEPLOY_EC2_CHANNEL").
				OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_CHANNEL").
				HintOptions("stable", "beta", "alpha").String()

	flEc2DeployEtcdToken = cmdEc2Deploy.Flag("etcd-token", "Etcd bootstrap token [ auto | <token> ]").
				Default("auto").OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_ETCD_TOKEN").
				Short('t').HintOptions("auto").String()

	flEc2DeployNs1ApiKey = cmdEc2Deploy.Flag("ns1-api-key", "NS1 private API key.").
				Required().PlaceHolder("KATO_DEPLOY_EC2_NS1_API_KEY").
				OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_NS1_API_KEY").
				String()

	flEc2DeployCaCert = cmdEc2Deploy.Flag("ca-cert", "Path to CA certificate.").
				PlaceHolder("KATO_DEPLOY_EC2_CA_CET").
				OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_CA_CET").
				Short('c').String()

	flEc2DeployRegion = cmdEc2Deploy.Flag("region", "Amazon EC2 region.").
				Required().PlaceHolder("KATO_DEPLOY_EC2_REGION").
				OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_REGION").
				Short('r').String()

	flEc2DeployZone = cmdEc2Deploy.Flag("zone", "Amazon EC2 availability zone.").
			Required().PlaceHolder("KATO_DEPLOY_EC2_ZONE").
			OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_ZONE").
			String()

	flEc2DeployDomain = cmdEc2Deploy.Flag("domain", "Used to identify the VPC.").
				Required().PlaceHolder("KATO_DEPLOY_EC2_DOMAIN").
				OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_DOMAIN").
				Short('d').String()

	flEc2DeployKeyPair = cmdEc2Deploy.Flag("key-pair", "EC2 key pair.").
				Required().PlaceHolder("KATO_DEPLOY_EC2_KEY_PAIR").
				OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_KEY_PAIR").
				Short('k').String()

	flEc2DeployVpcCidrBlock = cmdEc2Deploy.Flag("vpc-cidr-block", "IPs to be used by the VPC.").
				Default("10.0.0.0/16").OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_VPC_CIDR_BLOCK").
				String()

	flEc2DeployIntSubnetCidr = cmdEc2Deploy.Flag("internal-subnet-cidr", "CIDR for the internal subnet.").
					Default("10.0.1.0/24").OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_INTERNAL_SUBNET_CIDR").
					String()

	flEc2DeployExtSubnetCidr = cmdEc2Deploy.Flag("external-subnet-cidr", "CIDR for the external subnet.").
					Default("10.0.0.0/24").OverrideDefaultFromEnvar("KATO_DEPLOY_EC2_EXTERNAL_SUBNET_CIDR").
					String()

	flEc2DeployFlannelNetwork = cmdEc2Deploy.Flag("flannel-network", "Flannel entire overlay network.").
					Default("10.128.0.0/21").OverrideDefaultFromEnvar("KATO_DEPLOY_FLANNEL_NETWORK").
					String()

	flEc2DeployFlannelSubnetLen = cmdEc2Deploy.Flag("flannel-subnet-len", "Subnet len to llocate to each host.").
					Default("27").OverrideDefaultFromEnvar("KATO_DEPLOY_FLANNEL_SUBNET_LEN").
					String()

	flEc2DeployFlannelSubnetMin = cmdEc2Deploy.Flag("flannel-subnet-min", "Minimum subnet IP addresses.").
					Default("10.128.0.192").OverrideDefaultFromEnvar("KATO_DEPLOY_FLANNEL_SUBNET_MIN").
					String()

	flEc2DeployFlannelSubnetMax = cmdEc2Deploy.Flag("flannel-subnet-max", "Maximum subnet IP addresses.").
					Default("10.128.7.224").OverrideDefaultFromEnvar("KATO_DEPLOY_FLANNEL_SUBNET_MAX").
					String()

	flEc2DeployFlannelBackend = cmdEc2Deploy.Flag("flannel-backend", "Flannel backend type: [ udp | vxlan | host-gw | gce | aws-vpc | alloc ]").
					Default("vxlan").OverrideDefaultFromEnvar("KATO_DEPLOY_FLANNEL_BACKEND").
					HintOptions("udp", "vxlan", "host-gw", "gce", "aws-vpc", "alloc").String()

	//---------------------------
	// ec2 setup: nested command
	//---------------------------

	cmdEc2Setup = cmdEc2.Command("setup", "Setup an EC2 VPC and all the related components.")

	flEc2SetupClusterID = cmdEc2Setup.Flag("cluster-id", "Cluster ID for later reference.").
				Required().PlaceHolder("KATO_SETUP_EC2_CLUSTER_ID").
				OverrideDefaultFromEnvar("KATO_SETUP_EC2_CLUSTER_ID").
				String()

	flEc2SetupDomain = cmdEc2Setup.Flag("domain", "Used to identify the VPC..").
				Required().PlaceHolder("KATO_SETUP_EC2_DOMAIN").
				OverrideDefaultFromEnvar("KATO_SETUP_EC2_DOMAIN").
				Short('t').String()

	flEc2SetupRegion = cmdEc2Setup.Flag("region", "EC2 region.").
				Required().PlaceHolder("KATO_SETUP_EC2_REGION").
				OverrideDefaultFromEnvar("KATO_SETUP_EC2_REGION").
				Short('r').String()

	flEc2SetupZone = cmdEc2Setup.Flag("zone", "EC2 availability zone.").
			Required().PlaceHolder("KATO_SETUP_EC2_ZONE").
			OverrideDefaultFromEnvar("KATO_SETUP_EC2_ZONE").
			Short('z').String()

	flEc2SetupVpcCidrBlock = cmdEc2Setup.Flag("vpc-cidr-block", "IPs to be used by the VPC.").
				Default("10.0.0.0/16").OverrideDefaultFromEnvar("KATO_SETUP_EC2_VPC_CIDR_BLOCK").
				Short('c').String()

	flEc2SetupIntSubnetCidr = cmdEc2Setup.Flag("internal-subnet-cidr", "CIDR for the internal subnet.").
				Default("10.0.1.0/24").OverrideDefaultFromEnvar("KATO_SETUP_EC2_INTERNAL_SUBNET_CIDR").
				Short('i').String()

	flEc2SetupExtSubnetCidr = cmdEc2Setup.Flag("external-subnet-cidr", "CIDR for the external subnet.").
				Default("10.0.0.0/24").OverrideDefaultFromEnvar("KATO_SETUP_EC2_EXTERNAL_SUBNET_CIDR").
				Short('e').String()

	//-------------------------
	// ec2 add: nested command
	//-------------------------

	cmdEc2Add = cmdEc2.Command("add", "Adds a new instance to Káto.")

	flEc2AddRole = cmdEc2Add.Flag("role", "New instance role [ master | node | edge ]").
			Required().PlaceHolder("KATO_EC2_ADD_ROLE").
			OverrideDefaultFromEnvar("KATO_EC2_ADD_ROLE").
			HintOptions("master", "node", "edge").String()

	flEc2AddID = cmdEc2Add.Flag("id", "New instance ID.").
			Required().PlaceHolder("KATO_EC2_ADD_ID").
			OverrideDefaultFromEnvar("KATO_EC2_ADD_ID").
			String()

	//-------------------------
	// ec2 run: nested command
	//-------------------------

	cmdEc2Run = cmdEc2.Command("run", "Starts a CoreOS instance on Amazon EC2.")

	flEc2RunHostname = cmdEc2Run.Flag("hostname", "For the EC2 dashboard.").
				PlaceHolder("KATO_RUN_EC2_HOSTNAME").
				OverrideDefaultFromEnvar("KATO_RUN_EC2_HOSTNAME").
				Short('h').String()

	flEc2RunRegion = cmdEc2Run.Flag("region", "EC2 region.").
			Required().PlaceHolder("KATO_RUN_EC2_REGION").
			OverrideDefaultFromEnvar("KATO_RUN_EC2_REGION").
			Short('r').String()

	flEc2RunZone = cmdEc2Run.Flag("zone", "EC2 availability zone.").
			Required().PlaceHolder("KATO_RUN_EC2_ZONE").
			OverrideDefaultFromEnvar("KATO_RUN_EC2_ZONE").
			Short('z').String()

	flEc2RunImageID = cmdEc2Run.Flag("image-id", "EC2 image id.").
			Required().PlaceHolder("KATO_RUN_EC2_IMAGE_ID").
			OverrideDefaultFromEnvar("KATO_RUN_EC2_IMAGE_ID").
			Short('i').String()

	flEc2RunInsType = cmdEc2Run.Flag("instance-type", "EC2 instance type.").
			Required().PlaceHolder("KATO_RUN_EC2_INSTANCE_TYPE").
			OverrideDefaultFromEnvar("KATO_RUN_EC2_INSTANCE_TYPE").
			Short('t').String()

	flEc2RunKeyPair = cmdEc2Run.Flag("key-pair", "EC2 key pair.").
			Required().PlaceHolder("KATO_RUN_EC2_KEY_PAIR").
			OverrideDefaultFromEnvar("KATO_RUN_EC2_KEY_PAIR").
			Short('k').String()

	flEc2RunSubnetID = cmdEc2Run.Flag("subnet-id", "EC2 subnet ID.").
				Required().PlaceHolder("KATO_RUN_EC2_SUBNET_ID").
				OverrideDefaultFromEnvar("KATO_RUN_EC2_SUBNET_ID").
				String()

	flEc2RunSecGrpID = cmdEc2Run.Flag("security-group-id", "EC2 security group ID.").
				Required().PlaceHolder("KATO_RUN_EC2_SECURITY_GROUP_ID").
				OverrideDefaultFromEnvar("KATO_RUN_EC2_SECURITY_GROUP_ID").
				String()

	flEc2RunPublicIP = cmdEc2Run.Flag("public-ip", "Allocate a public IP [ true | false | elastic ]").
				Default("false").OverrideDefaultFromEnvar("KATO_RUN_EC2_PUBLIC_IP").
				HintOptions("true", "false", "elastic").String()

	flEc2RunIAMRole = cmdEc2Run.Flag("iam-role", "IAM role [ master | node | edge ]").
			OverrideDefaultFromEnvar("KATO_RUN_EC2_IAM_ROLE").
			HintOptions("master", "node", "edge").String()

	flEc2RunSrcDstCheck = cmdEc2Run.Flag("source-dest-check", " [ true | false ]").
				Default("true").OverrideDefaultFromEnvar("KATO_RUN_EC2_SOURCE_DEST_CHECK").
				HintOptions("true", "false").String()
)
