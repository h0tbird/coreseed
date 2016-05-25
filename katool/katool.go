package katool

//---------------------------------------------------------------------------
// Package factored import statement:
//---------------------------------------------------------------------------

import (

	// Stdlib:
	"encoding/json"
	"io/ioutil"
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

// EtcdToken takes masterCount and returns a valid etcd bootstrap token:
func EtcdToken(masterCount int) (string, error) {

	// Request an etcd bootstrap token:
	res, err := http.Get("https://discovery.etcd.io/new?size=" + strconv.Itoa(masterCount))
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
func LoadState(clusterID string) (map[string]interface{}, error) {

	// Load data from state file:
	stateFile := os.Getenv("HOME") + "/.kato/" + clusterID + ".json"
	raw, err := ioutil.ReadFile(stateFile)
	if err != nil {
		return nil, err
	}

	// Decode the loaded JSON data:
	var dat map[string]interface{}
	if err := json.Unmarshal(raw, &dat); err != nil {
		return nil, err
	}

	return dat, nil
}
