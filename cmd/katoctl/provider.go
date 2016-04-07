package main

//--------------------------------------------------------------------------
// Typedefs:
//--------------------------------------------------------------------------

type cloudProvider interface {
	Deploy() error
	Setup() error
	Run(udata []byte) error
}

//-------------------------------------------------------------------------
// func: deploy
//-------------------------------------------------------------------------

func deploy(cp cloudProvider) error {

	if err := cp.Deploy(); err != nil {
		return err
	}

	return nil
}

//-------------------------------------------------------------------------
// func: setup
//-------------------------------------------------------------------------

func setup(cp cloudProvider) error {

	if err := cp.Setup(); err != nil {
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
	if err = cp.Run(udata); err != nil {
		return err
	}

	return nil
}
