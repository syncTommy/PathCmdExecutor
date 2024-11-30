package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func runCommandInDir(dir string, cmd string) (string, error) {
	// 创建一个新的命令执行环境
	command := exec.Command("sh", "-c", cmd)
	// 设置工作目录
	command.Dir = dir
	// 设置标准输出和标准错误输出缓冲
	var out bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &out
	command.Stderr = &stderr
	// 执行命令
	err := command.Run()
	if err != nil {
		return "", fmt.Errorf("command failed: %v, stderr: %v", err, stderr.String())
	}
	// 返回输出
	return out.String(), nil
}

type dirCmdInfo struct {
	dir        string
	rawInfo    string
	simpleInfo string
	errInfo    string
}

type DirCmdRes struct {
	infos      []dirCmdInfo
	conclusion string
}

func RunCommandInDirBatchAndPrint(dirs []string, cmd string, timeout time.Duration) (err error) {

	res := RunCommandInDirBatchWithRetry(dirs, cmd, timeout)
	cmdStr := strings.Replace(cmd, " ", "-", -1)
	rawFileName := "./" + cmdStr + getCurrentDateTimeString() + "-raw" + ".log"
	simpleFileName := "./" + cmdStr + getCurrentDateTimeString() + "-simple" + ".log"
	errFileName := "./" + cmdStr + getCurrentDateTimeString() + "-err" + ".log"

	var rawStr string
	for _, inf := range res.infos {
		rawStr += inf.rawInfo
	}
	rawStr += res.conclusion
	err = createAndWriteFile(rawFileName, rawStr)
	if err != nil {
		return
	}

	var simpleStr string
	for _, inf := range res.infos {
		simpleStr += inf.simpleInfo
	}
	simpleStr += res.conclusion
	err = createAndWriteFile(simpleFileName, simpleStr)
	if err != nil {
		return
	}

	var errStr string
	for _, inf := range res.infos {
		errStr += inf.errInfo
	}
	errStr += res.conclusion
	err = createAndWriteFile(errFileName, errStr)
	return
}

func RunCommandInDirBatchWithRetry(dirs []string, cmd string, timeout time.Duration) DirCmdRes {
	start := time.Now()
	retryTimes := 0
	processingDirs := make([]string, len(dirs))

	copy(processingDirs, dirs)
	var totalInfoMap = make(map[string]dirCmdInfo, len(dirs))
	for {
		infos, _ := runCommandInDirBatch(processingDirs, cmd, timeout)
		retryTimes++
		for d, i := range infos {
			d := d
			i := i
			if i.errInfo == "" {
				totalInfoMap[d] = i
			}
		}
		var processAgain = make([]string, len(dirs))
		for _, dir := range processingDirs {
			if _, ok := totalInfoMap[dir]; !ok {
				processAgain = append(processAgain, dir)
			}
		}
		processingDirs = processAgain
		mustBreak := retryTimes >= TOTAL_RETRY_TIMES ||
			len(totalInfoMap) == len(dirs) ||
			time.Since(start) >= timeout

		if mustBreak {
			for d, i := range infos {
				d := d
				i := i
				totalInfoMap[d] = i
			}
			break
		}
	}
	for _, dir := range dirs {
		if _, ok := totalInfoMap[dir]; !ok {
			totalInfoMap[dir] = dirCmdInfo{
				dir:     dir,
				errInfo: "timeout error" + "\n",
			}
		}
	}
	// 排序

	res := DirCmdRes{}
	var notExecutedCnt, executedCnt, erroredCnt, succeededCnt int
	unfinishedStr := "=============== unfinished dirs (time out / retried enough) ===============\n"
	for _, dir := range dirs {
		if val, ok := totalInfoMap[dir]; !ok {
			unfinishedStr += dir + "\n"
			notExecutedCnt++
		} else {
			res.infos = append(res.infos, val)
			executedCnt++
			if val.errInfo != "" {
				erroredCnt++
			} else {
				succeededCnt++
			}
		}
	}
	unfinishedStr += "=============== unfinished dirs (time out / retried enough) ===============\n"

	timeCostStr := "=============== TIME COST ===============\n" +
		genTimeCostStr(start, time.Now()) +
		"\n=============== TIME COST ===============\n"

	res.conclusion += unfinishedStr
	res.conclusion += "****************************CONCLUSION****************************\n"
	res.conclusion += fmt.Sprintf("total dirs : %d =  notExecutedCnt : %d, executedCnt: %d (erroredCnt : %d + succeededCnt : %d )\n",
		len(dirs), notExecutedCnt, executedCnt, erroredCnt, succeededCnt)
	res.conclusion += "****************************CONCLUSION****************************\n"
	res.conclusion += timeCostStr

	return res
}

// 并发太大会出错
func runCommandInDirBatch(dirs []string, cmd string, timeout time.Duration) (infos map[string]dirCmdInfo, cost time.Duration) {
	start := time.Now()
	ticker := time.NewTicker(timeout).C

	var wg sync.WaitGroup
	wg.Add(len(dirs))
	var doneChan = make(chan struct{})
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	var cmdOutChan = make(chan dirCmdInfo, len(dirs))
	var cmdOutSlice = make([]dirCmdInfo, 0, len(dirs))

	//var stopProcess = make(chan struct{})

	var concurrency = CONCURRENCY
	var limitCh = make(chan struct{}, concurrency)

	for _, dir := range dirs {
		dir := dir

		go func() {
			defer func() {
				wg.Done()
				<-limitCh
			}()

			limitCh <- struct{}{}
			outputStr := cmd + "----->" + dir + ":\n"
			simpleStr := outputStr

			localOutput, localErr := runCommandInDir(dir, cmd)
			if localErr != nil {
				outputStr += localErr.Error()
				simpleStr += localErr.Error()
			} else {
				outputStr += localOutput
				simpleStr += extractBuildResult(localOutput)
			}
			outputStr += "\n================================================================================================\n"
			simpleStr += "\n================================================================================================\n"
			//select {
			//case <-stopProcess:
			//	return
			//default:
			//}
			info := dirCmdInfo{
				dir:        dir,
				rawInfo:    outputStr,
				simpleInfo: simpleStr,
			}
			if localErr != nil {
				info.errInfo = outputStr
			}
			cmdOutChan <- info

		}()
	}

	//var unfinished = false
lp:
	for {
		select {
		case <-ticker:
			break lp
		case <-doneChan:
			break lp
		}
	}
	//if unfinished {
	//	close(stopProcess)
	//}

	var wg2 sync.WaitGroup
	wg2.Add(1)
	go func() {
		for {
			select {
			case res := <-cmdOutChan:
				cmdOutSlice = append(cmdOutSlice, res)
			default:
				wg2.Done()
				return
			}
		}
	}()
	wg2.Wait()

	dir2InfoMap := make(map[string]dirCmdInfo)
	for _, info := range cmdOutSlice {
		dir2InfoMap[info.dir] = info
	}
	return dir2InfoMap, time.Now().Sub(start)
}
