package tools

//---------------------------------------------------------------------------
// Package factored import statement:
//---------------------------------------------------------------------------

import (

	// Stdlib:
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

//-----------------------------------------------------------------------------
// func: ExecutePipeline
//-----------------------------------------------------------------------------

// ExecutePipeline takes two commands and pipes the stdout of the first one
// into the stdin of the second one.
func ExecutePipeline(cmd1, cmd2 *exec.Cmd) error {

	var err error

	// Adjust the stderr:
	cmd1.Stderr = os.Stderr
	cmd2.Stderr = os.Stderr

	// Connect both commands:
	cmd2.Stdin, err = cmd1.StdoutPipe()
	if err != nil {
		return err
	}

	// Execute the pipeline:
	if err := cmd2.Start(); err != nil {
		return err
	}
	if err := cmd1.Run(); err != nil {
		return err
	}
	if err := cmd2.Wait(); err != nil {
		return err
	}

	// Return on success:
	return nil
}

//-----------------------------------------------------------------------------
// func: EtcdToken
//-----------------------------------------------------------------------------

// EtcdToken takes quorumCount and returns a valid etcd bootstrap token:
func EtcdToken(quorumCount int) (string, error) {

	// Request an etcd bootstrap token:
	res, err := http.Get("https://discovery.etcd.io/new?size=" + strconv.Itoa(quorumCount))
	if err != nil {
		return "", err
	}

	// Retrieve the token URL:
	tokenURL, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	// Call the close method:
	_ = res.Body.Close()

	// Return the token ID:
	slice := strings.Split(string(tokenURL), "/")
	return slice[len(slice)-1], nil
}

//-----------------------------------------------------------------------------
// func: LoadState
//-----------------------------------------------------------------------------

// LoadState reads the current ClusterID state file and decodes its content
// into a data structure:
func LoadState(clusterID string) ([]byte, error) {

	// Load data from state file:
	stateFile := os.Getenv("HOME") + "/.kato/" + clusterID + ".json"
	raw, err := ioutil.ReadFile(stateFile)
	if err != nil {
		return nil, err
	}

	return raw, nil
}

//-----------------------------------------------------------------------------
// func: OffsetIP
//-----------------------------------------------------------------------------

// OffsetIP takes a CIDR and an offset and returns the IP address at the offset
// position starting at the beginning of the CIDR's subnet:
func OffsetIP(cidr string, offset int) string {

	// Parse the CIDR:
	ip1, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return ""
	}

	// Compute the IP:
	ip2 := ip1.Mask(ipnet.Mask)
	a := int(ipToI32(ip2[len(ip2)-4:]))

	// Return:
	return i32ToIP(int32(a + offset)).String()
}

func ipToI32(ip net.IP) int32 {
	ip = ip.To4()
	return int32(ip[0])<<24 | int32(ip[1])<<16 | int32(ip[2])<<8 | int32(ip[3])
}

func i32ToIP(a int32) net.IP {
	return net.IPv4(byte(a>>24), byte(a>>16), byte(a>>8), byte(a))
}
