package main

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/process"

	"github.com/xyzj/gopsu"
)

func main() {
	// println(fmt.Sprintf("%+v", os.Args[1:]))
	var pidPoll sync.Map
	// 启动指定配置的 frpc
	var cmd *exec.Cmd
	for _, v := range os.Args[1:] {
		cmd = exec.Command(gopsu.JoinPathFromHere("frpc"), "-c", v)
		if err := cmd.Start(); err != nil {
			continue
		}
		println("--> start " + cmd.String() + " success")
		pidPoll.Store(int32(cmd.Process.Pid), cmd.String())
	}
	t := time.NewTicker(time.Second * 5)
	for range t.C {
		pidPoll.Range(func(key interface{}, value interface{}) bool {
			if ok, _ := process.PidExists(key.(int32)); ok {
				// println("--> " + cmd.String() + " seems good")
				return true
			}
			println("--> pid " + strconv.Itoa(int(key.(int32))) + " not found, try to restart...")
			scmd := strings.Split(value.(string), " ")
			cmd = exec.Command(scmd[0], scmd[1:]...)
			pidPoll.Store(int32(cmd.Process.Pid), cmd.String())
			pidPoll.Delete(key)
			return true
		})
	}
}
