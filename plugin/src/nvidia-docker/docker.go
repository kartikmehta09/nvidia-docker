// Copyright (c) 2015, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"syscall"
)

func docker(command string, arg ...string) ([]byte, error) {
	var buf bytes.Buffer

	args := append(append(DockerBin[1:], command), arg...)
	cmd := exec.Command(DockerBin[0], args...)
	cmd.Stderr = &buf

	b, err := cmd.Output()
	if err != nil {
		b = bytes.TrimSpace(buf.Bytes())
		return nil, fmt.Errorf("%s", b)
	}
	return b, nil
}

func DockerParseArgs(args []string, cmd ...string) (string, int, error) {
	type void struct{}

	re := regexp.MustCompile("(?m)^\\s*(-[^=]+)=[^{true}{false}].*$")
	flags := make(map[string]void)

	b, err := docker("help", cmd...)
	if err != nil {
		return "", -1, err
	}

	// Build the set of Docker flags taking an option using "docker help"
	for _, m := range re.FindAllSubmatch(b, -1) {
		for _, f := range bytes.Split(m[1], []byte(", ")) {
			flags[string(f)] = void{}
		}
	}
	for i := 0; i < len(args); i++ {
		if args[i][:1] == "-" {
			// Skip the flags and their options
			if _, ok := flags[args[i]]; ok {
				i++
			}
			continue
		}
		// Return the first arg that is not a flag
		return args[i], i, nil
	}
	return "", -1, nil
}

func DockerGetLabel(image, label string) (string, error) {
	format := fmt.Sprintf(`--format='{{index .Config.Labels "%s"}}'`, label)

	b, err := docker("inspect", format, image)
	if err != nil {
		return "", err
	}
	return string(bytes.Trim(b, " \n")), nil
}

func Docker(arg ...string) error {
	cmd, err := exec.LookPath(DockerBin[0])
	if err != nil {
		return err
	}
	args := append(DockerBin, arg...)

	return syscall.Exec(cmd, args, nil)
}
