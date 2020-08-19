package main

import (
	"os/exec"
	"time"

	"github.com/shirou/gopsutil/process"
)

func main() {
	var found bool
	for {
		select {
		case <-time.After(time.Minute):
			pl, err := process.Processes()
			if err != nil {
				println("get processes error: " + err.Error())
				continue
			}
			found = false
			for _, v := range pl {
				n, _ := v.Name()
				if n == "frpc" {
					s, _ := v.Status()
					if s != "S" {
						v.Kill()
						break
					}
					found = true
					break
				}
			}
			if !found {
				cmd := exec.Command("frpc", "-c", "frpc.ini")
				cmd.Dir = "/home/xy/bin/frp"
				cmd.Start()
			}
		}
	}
	// a, err := process.Processes()
	// if err != nil {
	// 	println(err.Error())
	// 	return
	// }
	// for _, v := range a {
	// 	n, _ := v.Name()
	// 	m, _ := v.Cmdline()
	// 	s, _ := v.Status()
	// 	if n == "frpc" {
	// 		println(v.Pid, n, m, s)
	// 	}
	// }
}
