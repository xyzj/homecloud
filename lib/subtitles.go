package lib

import (
	"bytes"
	"io/ioutil"
)

// Smi2Vtt smi转vtt
func Smi2Vtt(in, out string) error {
	bIn, err := ioutil.ReadFile(in)
	if err != nil {
		return err
	}

	var bOut bytes.Buffer
	bOut.WriteString("WEBVTT\r\n")

	return ioutil.WriteFile(out, bOut.Bytes(), 0644)
}
