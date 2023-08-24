package main

type Controller interface {
	Start(params ...string) error
	Stop() error
	Status() (status string, err error)
	List() (configsText string, err error)
}

type containerController struct {
	runtimeCmd string
	container  string
}

func NewContainerController(runtimeCmd, containerName string) Controller {
	cc := containerController{
		runtimeCmd: runtimeCmd,
		container:  containerName,
	}
	return cc
}

func (cc containerController) Start(params ...string) error {
	return nil
}

func (cc containerController) Status() (status string, err error) {
	return "", nil
}

func (cc containerController) Stop() error {
	return nil
}

func (cc containerController) List() (configsText string, err error) {
	return "", nil
}
