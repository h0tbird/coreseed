package kato

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
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
// func: DumpState
//-----------------------------------------------------------------------------

// DumpState serializes the given state as a clusterID JSON file.
func DumpState(s interface{}, clusterID string) error {

	// Marshal the data:
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	// Create the state directory:
	path := os.Getenv("HOME") + "/.kato"
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(path, 0700)
			if err != nil {
				return err
			}
		}
	}

	// Write the state file:
	err = ioutil.WriteFile(path+"/"+clusterID+".json", data, 0600)
	if err != nil {
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------
// func: ReadState
//-----------------------------------------------------------------------------

// ReadState reads the current ClusterID state file.
func ReadState(clusterID string) ([]byte, error) {

	// Read data from state file:
	stateFile := os.Getenv("HOME") + "/.kato/" + clusterID + ".json"
	raw, err := ioutil.ReadFile(stateFile)
	if err != nil {
		return nil, err
	}

	return raw, nil
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
func CreateDNSZones(wch *WaitChan, provider, apiKey, domain string) {

	// Decrement:
	defer wch.WaitGrp.Done()

	// Forge the zone command:
	cmd := exec.Command("katoctl", provider,
		"--api-key", apiKey, "zone", "add",
		domain, "int."+domain, "ext."+domain)

	// Execute the zone command:
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		wch.ErrChan <- err
	}
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
func NewEtcdToken(wch *WaitChan, quorumCount int, token *string) {

	// Decrement:
	defer wch.WaitGrp.Done()

	// Send the request:
	const etcdIO = "https://discovery.etcd.io/"
	res, err := http.Get(etcdIO + "new?size=" + strconv.Itoa(quorumCount))
	if err != nil {
		wch.ErrChan <- err
		return
	}

	// Get the response body:
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		wch.ErrChan <- err
		return
	}

	// Call the close method:
	if err := res.Body.Close(); err != nil {
		wch.ErrChan <- err
		return
	}

	// Test whether pattern matches string:
	match, err := regexp.MatchString(etcdIO+"([a-z,0-9]+$)", string(body))
	if err != nil {
		wch.ErrChan <- err
		return
	}

	// Return if invalid:
	if !match {
		wch.ErrChan <- errors.New("Invalid etcd token retrieved")
		return
	}

	// Return the token ID:
	slice := strings.Split(string(body), "/")
	*token = slice[len(slice)-1]
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
