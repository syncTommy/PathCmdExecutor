package main

import (
	"fmt"
	"time"
)

func main() {
	//discoverMavenProjects()
	execCmdBatch()

}

// 列举目录下的maven服务
func discoverMavenProjects() {
	dir := "/parent-dir-path"
	projects, err := findMavenProjects(dir)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, project := range projects {
		fmt.Println(project)
	}
}

func execCmd() {
	fmt.Println("exec cmd ...")
	dir := "/path"
	cmd := "mvn package"
	output, err := runCommandInDir(dir, cmd)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(output)
}

func execCmdBatch() {
	var dirs = []string{
		"/parent-dir-path",
	}
	for index := range dirs {
		dirs[index] = "/parent-path" + dirs[index]
	}
	cmd := "mvn clean "
	err := RunCommandInDirBatchAndPrint(dirs, cmd, time.Minute*15)
	if err != nil {
		fmt.Println("RunCommandInDirBatchAndPrint err:", err)
		return
	}
}
