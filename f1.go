package main

import "os"
import "path/filepath"

// 函数接收一个目录路径，返回一个字符串切片，包含所有包含 pom.xml 文件的子目录名。
func findMavenProjects(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var mavenProjects []string
	for _, file := range files {
		if file.IsDir() { // 检查子目录中是否存在 pom.xml 文件
			pomPath := filepath.Join(dir, file.Name(), "pom.xml")
			if _, err := os.Stat(pomPath); err == nil {
				mavenProjects = append(mavenProjects, file.Name())
			}
		}
	}
	return mavenProjects, nil
}
