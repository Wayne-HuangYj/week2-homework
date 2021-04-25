package main

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path"
	e "errors"
)

func ReadFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "open failed")
	}
	defer f.Close()
	return nil, nil
}


func ReadConfig() ([]byte, error) {
	home, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "no such file or directory")
	}
	config, err := ReadFile(path.Join(home, "sb.config"))
	return config, errors.WithMessage(err, "failed to read config")
}

var ErrNotFound = e.New("sb")
func myErrors() error {
	return ErrNotFound
}

func main() {
	_, err := ReadConfig()
	if err != nil {
		fmt.Println(errors.Cause(err))
		fmt.Printf("\n%+v\n", err)
		// os.Exit(1)
	}
	// err = myErrors()
	// fmt.Printf("\n%+v\n", err)
	fmt.Printf("%T, %v,\n", ErrNotFound, ErrNotFound == myErrors())
}