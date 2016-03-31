package main

//--------------------------------------------------------------------------
// Typedefs:
//--------------------------------------------------------------------------

type cloudProvider interface {
	Setup() error
	Run(udata []byte) error
}

//-------------------------------------------------------------------------
// func: setup
//-------------------------------------------------------------------------

func setup(cp cloudProvider) error {

	err := cp.Setup()
	if err != nil {
		return err
	}

	return nil
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
