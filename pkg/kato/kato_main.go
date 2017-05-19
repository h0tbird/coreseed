package kato

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

//-----------------------------------------------------------------------------
// WaitChan stuff:
//-----------------------------------------------------------------------------

// WaitChan is used to handle errors that occur in some goroutines.
type WaitChan struct {
	WaitGrp sync.WaitGroup
	ErrChan chan error
	EndChan chan bool
}

// NewWaitChan initializes a WaitChan struct.
func NewWaitChan(len int) *WaitChan {
	wch := new(WaitChan)
	wch.WaitGrp.Add(len)
	wch.ErrChan = make(chan error, 1)
	wch.EndChan = make(chan bool, 1)
	return wch
}

// WaitErr waits for any error or for all go routines to finish.
func (wch *WaitChan) WaitErr() error {

	// Put the wait group in a go routine:
	go func() {
		wch.WaitGrp.Wait()
		wch.EndChan <- true
	}()

	// This select will block:
	select {
	case <-wch.EndChan:
		return nil
	case err := <-wch.ErrChan:
		return err
	}
}

//-----------------------------------------------------------------------------
// func: CountNodes
//-----------------------------------------------------------------------------

// CountNodes returns the count of <role> nodes defined in <quads>.
func CountNodes(quads []string, role string) (count int) {

	// Default to zero:
	count = 0

	// Get the role count:
	for _, q := range quads {
		if strings.Contains(q, role) {
			s := strings.Split(q, ":")
			count, _ = strconv.Atoi(s[0])
			break
		}
	}

	return
}

//-----------------------------------------------------------------------------
// func: CreateDNSZones
//-----------------------------------------------------------------------------

// CreateDNSZones creates (int|ext).<domain> zones using <provider>.
func CreateDNSZones(wch *WaitChan, provider, apiKey, domain string) error {

	// Decrement:
	if wch != nil {
		defer wch.WaitGrp.Done()
	}

	// Forge the zone command:
	cmd := exec.Command("katoctl", provider,
		"--api-key", apiKey, "zone", "add",
		domain, "int."+domain, "ext."+domain)

	// Execute the zone command:
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if wch != nil {
			wch.ErrChan <- err
		}
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: ExecutePipeline
//-----------------------------------------------------------------------------

// ExecutePipeline takes two commands and pipes the stdout of the first one
// into the stdin of the second one. Returns the output as []byte.
func ExecutePipeline(cmd1, cmd2 *exec.Cmd) ([]byte, error) {

	var err error

	// Adjust the stderr:
	cmd1.Stderr = os.Stderr
	cmd2.Stderr = os.Stderr

	// Connect both commands:
	cmd2.Stdin, err = cmd1.StdoutPipe()
	if err != nil {
		return nil, err
	}

	// Get cmd2 stdout:
	stdout, err := cmd2.StdoutPipe()
	if err != nil {
		return nil, err
	}

	// Execute the pipeline:
	if err = cmd2.Start(); err != nil {
		return nil, err
	}
	if err = cmd1.Run(); err != nil {
		return nil, err
	}

	// Read the cmd2 output:
	out, err := ioutil.ReadAll(stdout)
	if err != nil {
		return nil, err
	}

	// Wait and return:
	if err = cmd2.Wait(); err != nil {
		return nil, err
	}

	return out, nil
}

//-----------------------------------------------------------------------------
// func: NewEtcdToken
//-----------------------------------------------------------------------------

// NewEtcdToken takes quorumCount and returns a valid etcd bootstrap token:
func NewEtcdToken(wch *WaitChan, quorumCount int, token *string) error {

	// Decrement:
	if wch != nil {
		defer wch.WaitGrp.Done()
	}

	// Request an etcd bootstrap token:
	res, err := http.Get("https://discovery.etcd.io/new?size=" + strconv.Itoa(quorumCount))
	if err != nil {
		if wch != nil {
			wch.ErrChan <- err
		}
		return err
	}

	// Retrieve the token URL:
	tokenURL, err := ioutil.ReadAll(res.Body)
	if err != nil {
		if wch != nil {
			wch.ErrChan <- err
		}
		return err
	}

	// Call the close method:
	_ = res.Body.Close()

	// Return the token ID:
	slice := strings.Split(string(tokenURL), "/")
	*token = slice[len(slice)-1]
	return nil
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
