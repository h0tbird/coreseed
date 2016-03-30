package main

//--------------------------------------------------------------------------
// Typedefs:
//--------------------------------------------------------------------------

type cloudProvider interface {
	Run(udata []byte) error
}

//--------------------------------------------------------------------------
// func: run
//--------------------------------------------------------------------------

func run(cp cloudProvider) error {

	// Retrieve user data:
	udata, err := readUdata()
	if err != nil {
		return err
	}

	// Create the machine:
	err = cp.Run(udata)
	if err != nil {
		return err
	}

	return nil
}
