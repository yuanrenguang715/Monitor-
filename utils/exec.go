package utils

import (
	"bytes"
	"os/exec"
)

//ExecCommand 命令
func ExecCommand(comm []string) (string, error) {
	cmd := exec.Command(comm[0], comm[1:]...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

// ExecCommandSs 获取端口连接数
/*func ExecCommandSs(port string) ([]string, error) {
	cmd := exec.Command("ss", "state", "all", "sport", "eq", port)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return PerrAddress(out.String()), nil
}*/
