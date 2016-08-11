package main

//-----------------------------------------------------------------------------
// 'katoctl ec2' command flags definitions:
//-----------------------------------------------------------------------------

var (

	//------------------------------------
	// List of EC2 instances and regions:
	//------------------------------------

	ec2Instances = []string{
		"c3.2xlarge", "c3.4xlarge", "c3.8xlarge", "c3.large", "c3.xlarge", "cc2.8xlarge",
		"cg1.4xlarge", "d2.2xlarge", "d2.4xlarge", "d2.8xlarge", "d2.xlarge", "g2.2xlarge",
		"g2.8xlarge", "hi1.4xlarge", "hs1.8xlarge", "i2.2xlarge", "i2.4xlarge", "i2.8xlarge",
		"i2.xlarge", "m3.2xlarge", "m3.large", "m3.medium", "m3.xlarge", "r3.2xlarge",
		"r3.4xlarge", "r3.8xlarge", "r3.large", "r3.xlarge", "x1.32xlarge"}

	ec2Regions = []string{
		"us-east-1", "us-west-1", "us-west-2", "eu-west-1", "eu-central-1", "ap-northeast-1",
		"ap-northeast-2", "ap-southeast-1", "ap-southeast-2", "sa-east-1"}

	//------------------------
	// ec2: top level command
	//------------------------

	cmdEc2 = app.Command("ec2", "This is the Káto EC2 provider.")

	//----------------------------
	// ec2 deploy: nested command
	//----------------------------

	cmdEc2Deploy = cmdEc2.Command("deploy",
		"Deploy Káto's infrastructure on Amazon EC2.")

	flEc2DeployClusterID = regexpMatch(cmdEc2Deploy.Flag("cluster-id",
		"Cluster ID for later reference.").
		Required().PlaceHolder("KATO_EC2_DEPLOY_CLUSTER_ID").
		OverrideDefaultFromEnvar("KATO_EC2_DEPLOY_CLUSTER_ID"), "^[a-zA-Z0-9-]+$")

	flEc2DeployQuorumCount = cmdEc2Deploy.Flag("quorum-count",
		"Number of quorum nodes in the cluster [ 1 | 3 | 5 ]").
		Required().PlaceHolder("KATO_EC2_DEPLOY_QUORUM_COUNT").
		OverrideDefaultFromEnvar("KATO_EC2_DEPLOY_QUORUM_COUNT").
		HintOptions("1", "3", "5").Int()

	flEc2DeployChannel = cmdEc2Deploy.Flag("channel",
		"CoreOS release channel [ stable | beta | alpha ]").
		Default("stable").PlaceHolder("KATO_EC2_DEPLOY_CHANNEL").
		OverrideDefaultFromEnvar("KATO_EC2_DEPLOY_CHANNEL").
		Enum("stable", "beta", "alpha")

	flEc2DeployEtcdToken = cmdEc2Deploy.Flag("etcd-token",
		"Etcd bootstrap token [ auto | <token> ]").
		Default("auto").OverrideDefaultFromEnvar("KATO_EC2_DEPLOY_ETCD_TOKEN").
		HintOptions("auto").String()

	flEc2DeployNs1ApiKey = cmdEc2Deploy.Flag("ns1-api-key",
		"NS1 private API key.").
		Required().PlaceHolder("KATO_EC2_DEPLOY_NS1_API_KEY").
		OverrideDefaultFromEnvar("KATO_EC2_DEPLOY_NS1_API_KEY").
		String()

	flEc2DeployCaCert = cmdEc2Deploy.Flag("ca-cert",
		"Path to CA certificate.").
		PlaceHolder("KATO_EC2_DEPLOY_CA_CET").
		OverrideDefaultFromEnvar("KATO_EC2_DEPLOY_CA_CET").
		ExistingFile()

	flEc2DeployRegion = cmdEc2Deploy.Flag("region",
		"Amazon EC2 region.").
		Required().PlaceHolder("KATO_EC2_DEPLOY_REGION").
		OverrideDefaultFromEnvar("KATO_EC2_DEPLOY_REGION").
		Enum(ec2Regions...)

	flEc2DeployZone = cmdEc2Deploy.Flag("zone",
		"Amazon EC2 availability zone.").
		Default("a").PlaceHolder("KATO_EC2_DEPLOY_ZONE").
		OverrideDefaultFromEnvar("KATO_EC2_DEPLOY_ZONE").
		Enum("a", "b", "c", "d")

	flEc2DeployDomain = cmdEc2Deploy.Flag("domain",
		"Used to identify the VPC.").
		Required().PlaceHolder("KATO_EC2_DEPLOY_DOMAIN").
		OverrideDefaultFromEnvar("KATO_EC2_DEPLOY_DOMAIN").
		String()

	flEc2DeployKeyPair = cmdEc2Deploy.Flag("key-pair",
		"EC2 key pair.").
		Required().PlaceHolder("KATO_EC2_DEPLOY_KEY_PAIR").
		OverrideDefaultFromEnvar("KATO_EC2_DEPLOY_KEY_PAIR").
		String()

	flEc2DeployVpcCidrBlock = cmdEc2Deploy.Flag("vpc-cidr-block",
		"IPs to be used by the VPC.").
		Default("10.0.0.0/16").
		OverrideDefaultFromEnvar("KATO_EC2_DEPLOY_VPC_CIDR_BLOCK").
		String()

	flEc2DeployIntSubnetCidr = cmdEc2Deploy.Flag("internal-subnet-cidr",
		"CIDR for the internal subnet.").
		Default("10.0.1.0/24").
		OverrideDefaultFromEnvar("KATO_EC2_DEPLOY_INTERNAL_SUBNET_CIDR").
		String()

	flEc2DeployExtSubnetCidr = cmdEc2Deploy.Flag("external-subnet-cidr",
		"CIDR for the external subnet.").
		Default("10.0.0.0/24").
		OverrideDefaultFromEnvar("KATO_EC2_DEPLOY_EXTERNAL_SUBNET_CIDR").
		String()

	flEc2DeployFlannelNetwork = cmdEc2Deploy.Flag("flannel-network",
		"Flannel entire overlay network.").
		Default("10.128.0.0/21").
		OverrideDefaultFromEnvar("KATO_EC2_DEPLOY_FLANNEL_NETWORK").
		String()

	flEc2DeployFlannelSubnetLen = cmdEc2Deploy.Flag("flannel-subnet-len",
		"Subnet len to llocate to each host.").
		Default("27").OverrideDefaultFromEnvar("KATO_EC2_DEPLOY_FLANNEL_SUBNET_LEN").
		String()

	flEc2DeployFlannelSubnetMin = cmdEc2Deploy.Flag("flannel-subnet-min",
		"Minimum subnet IP addresses.").
		Default("10.128.0.192").
		OverrideDefaultFromEnvar("KATO_EC2_DEPLOY_FLANNEL_SUBNET_MIN").
		String()

	flEc2DeployFlannelSubnetMax = cmdEc2Deploy.Flag("flannel-subnet-max",
		"Maximum subnet IP addresses.").
		Default("10.128.7.224").
		OverrideDefaultFromEnvar("KATO_EC2_DEPLOY_FLANNEL_SUBNET_MAX").
		String()

	flEc2DeployFlannelBackend = cmdEc2Deploy.Flag("flannel-backend",
		"Flannel backend type: [ udp | vxlan | host-gw | gce | aws-vpc | alloc ]").
		Default("vxlan").OverrideDefaultFromEnvar("KATO_EC2_DEPLOY_FLANNEL_BACKEND").
		Enum("udp", "vxlan", "host-gw", "gce", "aws-vpc", "alloc")

	arEc2DeployQuadruplet = cmdEc2Deploy.Arg("quadruplet",
		"<number_of_instances>:<instance_type>:<host_name>:<comma_separated_list_of_roles>").
		Required().Strings()

	//---------------------------
	// ec2 setup: nested command
	//---------------------------

	cmdEc2Setup = cmdEc2.Command("setup",
		"Setup IAM, VPC and EC2 components.")

	flEc2SetupClusterID = regexpMatch(cmdEc2Setup.Flag("cluster-id",
		"Cluster ID for later reference.").
		Required().PlaceHolder("KATO_EC2_SETUP_CLUSTER_ID").
		OverrideDefaultFromEnvar("KATO_EC2_SETUP_CLUSTER_ID"), "^[a-zA-Z0-9-]+$")

	flEc2SetupDomain = cmdEc2Setup.Flag("domain",
		"Used to identify the VPC..").
		Required().PlaceHolder("KATO_EC2_SETUP_DOMAIN").
		OverrideDefaultFromEnvar("KATO_EC2_SETUP_DOMAIN").
		String()

	flEc2SetupRegion = cmdEc2Setup.Flag("region",
		"EC2 region.").
		Required().PlaceHolder("KATO_EC2_SETUP_REGION").
		OverrideDefaultFromEnvar("KATO_EC2_SETUP_REGION").
		Enum(ec2Regions...)

	flEc2SetupZone = cmdEc2Setup.Flag("zone",
		"EC2 availability zone.").
		Default("a").PlaceHolder("KATO_EC2_SETUP_ZONE").
		OverrideDefaultFromEnvar("KATO_EC2_SETUP_ZONE").
		Enum("a", "b", "c", "d")

	flEc2SetupVpcCidrBlock = cmdEc2Setup.Flag("vpc-cidr-block",
		"IPs to be used by the VPC.").
		Default("10.0.0.0/16").
		OverrideDefaultFromEnvar("KATO_EC2_SETUP_VPC_CIDR_BLOCK").
		String()

	flEc2SetupIntSubnetCidr = cmdEc2Setup.Flag("internal-subnet-cidr",
		"CIDR for the internal subnet.").
		Default("10.0.1.0/24").
		OverrideDefaultFromEnvar("KATO_EC2_SETUP_INTERNAL_SUBNET_CIDR").
		String()

	flEc2SetupExtSubnetCidr = cmdEc2Setup.Flag("external-subnet-cidr",
		"CIDR for the external subnet.").
		Default("10.0.0.0/24").
		OverrideDefaultFromEnvar("KATO_EC2_SETUP_EXTERNAL_SUBNET_CIDR").
		String()

	//-------------------------
	// ec2 add: nested command
	//-------------------------

	cmdEc2Add = cmdEc2.Command("add",
		"Adds a new instance to an existing Káto cluster on EC2.")

	flEc2AddCluserID = regexpMatch(cmdEc2Add.Flag("cluster-id",
		"Cluster ID").
		Required().PlaceHolder("KATO_EC2_ADD_CLUSTER_ID").
		OverrideDefaultFromEnvar("KATO_EC2_ADD_CLUSTER_ID"), "^[a-zA-Z0-9-]+$")

	flEc2AddRoles = cmdEc2Add.Flag("roles",
		"Comma separated list of roles [ quorum | master | worker | border ]").
		Required().PlaceHolder("KATO_EC2_ADD_ROLES").
		OverrideDefaultFromEnvar("KATO_EC2_ADD_ROLES").
		String()

	flEc2AddHostName = cmdEc2Add.Flag("host-name",
		"hostname = <host-name>-<host-id>").
		Required().PlaceHolder("KATO_EC2_ADD_HOST_NAME").
		OverrideDefaultFromEnvar("KATO_EC2_ADD_HOST_NAME").
		String()

	flEc2AddHostID = cmdEc2Add.Flag("host-id",
		"hostname = <host-name>-<host-id>").
		Required().PlaceHolder("KATO_EC2_ADD_HOST_ID").
		OverrideDefaultFromEnvar("KATO_EC2_ADD_HOST_ID").
		String()

	flEc2AddAmiID = cmdEc2Add.Flag("ami-id",
		"CoreOS Amazon AMI ID to use.").
		PlaceHolder("KATO_EC2_ADD_AMI_ID").
		OverrideDefaultFromEnvar("KATO_EC2_ADD_AMI_ID").
		String()

	flEc2AddInsanceType = cmdEc2Add.Flag("instance-type",
		"EC2 instance type.").
		Required().PlaceHolder("KATO_EC2_ADD_INSTANCE_TYPE").
		OverrideDefaultFromEnvar("KATO_EC2_ADD_INSTANCE_TYPE").
		Enum(ec2Instances...)

	//-------------------------
	// ec2 run: nested command
	//-------------------------

	cmdEc2Run = cmdEc2.Command("run", "Starts a CoreOS instance on Amazon EC2.")

	flEc2RunTagName = cmdEc2Run.Flag("tag-name",
		"Tag name for the EC2 dashboard.").
		PlaceHolder("KATO_EC2_RUN_TAG_NAME").
		OverrideDefaultFromEnvar("KATO_EC2_RUN_TAG_NAME").
		String()

	flEc2RunRegion = cmdEc2Run.Flag("region",
		"EC2 region.").
		Required().PlaceHolder("KATO_EC2_RUN_REGION").
		OverrideDefaultFromEnvar("KATO_EC2_RUN_REGION").
		Enum(ec2Regions...)

	flEc2RunZone = cmdEc2Run.Flag("zone",
		"EC2 availability zone.").
		Default("a").PlaceHolder("KATO_EC2_RUN_ZONE").
		OverrideDefaultFromEnvar("KATO_EC2_RUN_ZONE").
		Enum("a", "b", "c", "d")

	flEc2RunAmiID = cmdEc2Run.Flag("ami-id",
		"EC2 AMI ID.").
		Required().PlaceHolder("KATO_EC2_RUN_AMI_ID").
		OverrideDefaultFromEnvar("KATO_EC2_RUN_AMI_ID").
		String()

	flEc2RunInstanceType = cmdEc2Run.Flag("instance-type",
		"EC2 instance type.").
		Required().PlaceHolder("KATO_EC2_RUN_INSTANCE_TYPE").
		OverrideDefaultFromEnvar("KATO_EC2_RUN_INSTANCE_TYPE").
		Enum(ec2Instances...)

	flEc2RunKeyPair = cmdEc2Run.Flag("key-pair",
		"EC2 key pair.").
		Required().PlaceHolder("KATO_EC2_RUN_KEY_PAIR").
		OverrideDefaultFromEnvar("KATO_EC2_RUN_KEY_PAIR").
		String()

	flEc2RunSubnetID = cmdEc2Run.Flag("subnet-id",
		"EC2 subnet ID.").
		Required().PlaceHolder("KATO_EC2_RUN_SUBNET_ID").
		OverrideDefaultFromEnvar("KATO_EC2_RUN_SUBNET_ID").
		String()

	flEc2RunSecGrpID = cmdEc2Run.Flag("security-group-id",
		"EC2 security group ID.").
		Required().PlaceHolder("KATO_EC2_RUN_SECURITY_GROUP_ID").
		OverrideDefaultFromEnvar("KATO_EC2_RUN_SECURITY_GROUP_ID").
		String()

	flEc2RunPublicIP = cmdEc2Run.Flag("public-ip",
		"Allocate a public IP [ true | false | elastic ]").
		Default("false").OverrideDefaultFromEnvar("KATO_EC2_RUN_PUBLIC_IP").
		Enum("true", "false", "elastic")

	flEc2RunIAMRole = cmdEc2Run.Flag("iam-role",
		"IAM role [ master | worker | border ]").
		OverrideDefaultFromEnvar("KATO_EC2_RUN_IAM_ROLE").
		Enum("master", "worker", "border")

	flEc2RunSrcDstCheck = cmdEc2Run.Flag("source-dest-check",
		" [ true | false ]").
		Default("true").OverrideDefaultFromEnvar("KATO_EC2_RUN_SOURCE_DEST_CHECK").
		Enum("true", "false")

	flEc2RunELBName = regexpMatch(cmdEc2Run.Flag("elb-name",
		"Register with existing ELB by name").
		OverrideDefaultFromEnvar("KATO_EC2_RUN_ELB_NAME"), "^[a-zA-Z0-9-]+$")

	flEc2RunPrivateIP = cmdEc2Run.Flag("private-ip",
		"The private IP address of the network interface.").
		OverrideDefaultFromEnvar("KATO_EC2_RUN_PRIVATE_IP").String()
)
