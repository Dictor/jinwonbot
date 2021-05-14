package main

import "os"

var ysArt string

func readYsArt() error {
	data, err := os.ReadFile("art.txt")
	if err != nil {
		return err
	}
	ysArt = string(data)
	return nil
}
