package katoctl

//-----------------------------------------------------------------------------
// Import:
//-----------------------------------------------------------------------------

import "github.com/katosys/kato/pkg/cli"

//-----------------------------------------------------------------------------
// 'katoctl udata' command flags definitions:
//-----------------------------------------------------------------------------

var (

	//--------------------------
	// udata: top level command
	//--------------------------

	cmdUdata = cli.App.Command("udata", "Generate CoreOS cloud-config user-data.")

	flUdataQuorumCount = cmdUdata.Flag("quorum-count",
		"Number of initial quorum nodes [ 1 | 3 | 5 ]").
		Required().OverrideDefaultFromEnvar("KATO_UDATA_QUORUM_COUNT").
		HintOptions("1", "3", "5").Int()

	flUdataMasterCount = cmdUdata.Flag("master-count",
		"Number of initial master nodes").
		Required().OverrideDefaultFromEnvar("KATO_UDATA_MASTER_COUNT").
		Int()

	flUdataClusterID = cmdUdata.Flag("cluster-id",
		"Cluster ID.").
		Required().PlaceHolder("KATO_UDATA_CLUSTER_ID").
		OverrideDefaultFromEnvar("KATO_UDATA_CLUSTER_ID").
		String()

	flUdataClusterState = cmdUdata.Flag("cluster-state",
		"Initial cluster state [ new | existing ]").
		Default("existing").PlaceHolder("KATO_UDATA_CLUSTER_STATE").
		OverrideDefaultFromEnvar("KATO_UDATA_CLUSTER_STATE").
		HintOptions("new", "existing").String()

	flUdataHostName = cmdUdata.Flag("host-name",
		"hostname = <host-name>-<host-id>").
		Required().PlaceHolder("KATO_UDATA_HOST_NAME").
		OverrideDefaultFromEnvar("KATO_UDATA_HOST_NAME").
		String()

	flUdataHostID = cmdUdata.Flag("host-id",
		"Must be a number: hostname = <host-name>-<host-id>").
		Required().PlaceHolder("KATO_UDATA_HOST_ID").
		OverrideDefaultFromEnvar("KATO_UDATA_HOST_ID").
		String()

	flUdataDomain = cmdUdata.Flag("domain",
		"Domain name as in (hostname -d)").
		Required().PlaceHolder("KATO_UDATA_DOMAIN").
		OverrideDefaultFromEnvar("KATO_UDATA_DOMAIN").
		String()

	flUdataRoles = cmdUdata.Flag("roles",
		"Comma separated list of roles [ quorum | master | worker | border ]").
		Required().PlaceHolder("KATO_UDATA_ROLES").
		OverrideDefaultFromEnvar("KATO_UDATA_ROLES").
		String()

	flUdataNs1Apikey = cmdUdata.Flag("ns1-api-key",
		"NS1 private API key.").
		PlaceHolder("KATO_UDATA_NS1_API_KEY").
		OverrideDefaultFromEnvar("KATO_UDATA_NS1_API_KEY").
		String()

	flUdataCaCertPath = cmdUdata.Flag("ca-cert-path",
		"Path to CA certificate.").
		PlaceHolder("KATO_UDATA_CA_CERT_PATH").
		OverrideDefaultFromEnvar("KATO_UDATA_CA_CERT_PATH").
		ExistingFile()

	flUdataEtcdToken = cmdUdata.Flag("etcd-token",
		"Provide an etcd discovery token.").
		PlaceHolder("KATO_UDATA_ETCD_TOKEN").
		OverrideDefaultFromEnvar("KATO_UDATA_ETCD_TOKEN").
		String()

	flUdataGzipUdata = cmdUdata.Flag("gzip-udata",
		"Enable udata compression.").
		Default("false").OverrideDefaultFromEnvar("KATO_UDATA_GZIP_UDATA").
		Bool()

	flUdataCalicoIPPool = cmdUdata.Flag("calico-ip-pool",
		"IP pool from which Calico expects endpoint IPs to be assigned.").
		Default("10.128.0.0/21").
		OverrideDefaultFromEnvar("KATO_UDATA_CALICO_IP_POOL").
		String()

	flUdataRexrayStorageDriver = cmdUdata.Flag("rexray-storage-driver",
		"REX-Ray storage driver: [ ebs | virtualbox ]").
		PlaceHolder("KATO_UDATA_REXRAY_STORAGE_DRIVER").
		OverrideDefaultFromEnvar("KATO_UDATA_REXRAY_STORAGE_DRIVER").
		Enum("virtualbox", "ebs")

	flUdataRexrayEndpointIP = cmdUdata.Flag("rexray-endpoint-ip",
		"REX-Ray endpoint IP address.").
		PlaceHolder("KATO_UDATA_REXRAY_ENDPOINT_IP").
		OverrideDefaultFromEnvar("KATO_UDATA_REXRAY_ENDPOINT_IP").
		String()

	flUdataEc2Region = cmdUdata.Flag("ec2-region",
		"EC2 region.").
		Default("eu-west-1").PlaceHolder("KATO_UDATA_EC2_REGION").
		OverrideDefaultFromEnvar("KATO_UDATA_EC2_REGION").
		Enum(ec2Regions...)

	flUdataIaasProvider = cmdUdata.Flag("iaas-provider",
		"IaaS provider [ vbox | ec2 | pkt ]").
		Required().PlaceHolder("KATO_UDATA_IAAS_PROVIDER").
		OverrideDefaultFromEnvar("KATO_UDATA_IAAS_PROVIDER").
		Enum("vbox", "ec2", "pkt")

	flUdataSlackWebhook = cmdUdata.Flag("slack-webhook",
		"Slack webhook URL.").
		PlaceHolder("KATO_UDATA_SLACK_WEBHOOK").
		OverrideDefaultFromEnvar("KATO_UDATA_SLACK_WEBHOOK").
		String()

	flUdataSysdigAccessKey = cmdUdata.Flag("sysdig-access-key",
		"SysDig secret access key.").
		PlaceHolder("KATO_UDATA_SYSDIG_ACCESS_KEY").
		OverrideDefaultFromEnvar("KATO_UDATA_SYSDIG_ACCESS_KEY").
		String()

	flUdataDatadogAPIKey = cmdUdata.Flag("datadog-api-key",
		"Datadog secret API key.").
		PlaceHolder("KATO_UDATA_DATADOG_API_KEY").
		OverrideDefaultFromEnvar("KATO_UDATA_DATADOG_API_KEY").
		String()

	flUdataStubZones = cmdUdata.Flag("stub-zone",
		"Use different nameservers for given domains.").
		PlaceHolder("KATO_UDATA_STUB_ZONE").
		OverrideDefaultFromEnvar("KATO_UDATA_STUB_ZONE").
		Strings()

	flUdataPrometheus = cmdUdata.Flag("prometheus",
		"Enable prometheus and/or its exporters.").
		Default("false").OverrideDefaultFromEnvar("KATO_UDATA_PROMETHEUS").
		Bool()

	flUdataSMTPURL = regexpMatch(cmdUdata.Flag("smtp-url",
		"SMTP server URL: <smtp://user:pass@host:port>").
		PlaceHolder("KATO_UDATA_SMTP_URL").
		OverrideDefaultFromEnvar("KATO_UDATA_SMTP_URL"), "^smtp://(.+):(.+)@(.+):(\\d+)$")

	flUdataAdminEmail = regexpMatch(cmdUdata.Flag("admin-email",
		"Administrator e-mail for cluster notifications.").
		PlaceHolder("KATO_UDATA_ADMIN_EMAIL").
		OverrideDefaultFromEnvar("KATO_UDATA_ADMIN_EMAIL"), "^[\\w-.+]+@[\\w-.+]+\\.[a-z]{2,4}$")
)
