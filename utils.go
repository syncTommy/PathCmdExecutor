package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	mvn_build_failure = "BUILD FAILURE"
	mvn_build_success = "BUILD SUCCESS"
	mvn_unknown       = "MVN UNKNOWN"
)

const (
	CONCURRENCY       = 5
	TOTAL_RETRY_TIMES = 5
)

func extractBuildResult(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, mvn_build_failure) {
			return mvn_build_failure
		}
		if strings.Contains(line, mvn_build_success) {
			return mvn_build_success
		}
	}
	return mvn_unknown
}

func getCurrentDateTimeString() string {
	return time.Now().Format("2006-01-02-15-04-05")
}

func genTimeCostStr(start, end time.Time) string {
	timeCostSecTotal := int64(end.Sub(start).Seconds())
	timeCostMinute := timeCostSecTotal / 60
	timeCostSecRest := timeCostSecTotal % 60
	timeCostStr := fmt.Sprintf("TIMECOST: %d minutes, %d seconds", timeCostMinute, timeCostSecRest)
	return timeCostStr
}

func createAndWriteFile(outputFileName, output string) error {
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		fmt.Println("create file "+outputFileName+" error", err)
		return err
	}
	_, err = outputFile.WriteString(output)
	if err != nil {
		fmt.Println("write "+outputFileName+" error", err)
		return err
	}
	outputFile.Close()
	return nil
}
