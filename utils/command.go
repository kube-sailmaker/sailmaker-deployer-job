package utils

import (
	"bufio"
	"bytes"
	"log"
	"os/exec"
	"strings"
)

func ExecuteAndDisplay(cmdName string, args []string) error {
	output, err := run(cmdName, args)
	if err != nil {
		log.Printf("error %v", err)
		log.Printf(strings.Join(output, "\n"))
		return err
	}
	log.Println(strings.Join(output, "\n"))
	return nil
}

func run(cmdName string, args []string) ([]string, error) {
	output := bytes.Buffer{}
	stderr := bytes.Buffer{}
	cmdRef := exec.Command(cmdName, args...)
	cmdRef.Stderr = &stderr
	cmdRef.Stdout = &output
	runErr := cmdRef.Run()
	if runErr != nil {
		return buffToLines(stderr), runErr
	}
	cmdRef.Wait()
	lines := buffToLines(output)
	return lines, nil
}

func buffToLines(bufferHandle bytes.Buffer) []string {
	scanner := bufio.NewScanner(&bufferHandle)
	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}


