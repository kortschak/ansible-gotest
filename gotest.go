// Copyright ©2020 Dan Kortschak. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// gotest is an ansible binary module that runs go test with test cache awareness.
// See https://docs.ansible.com/ansible/latest/dev_guide/developing_program_flow_modules.html#binary-modules.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		exit(response{
			Msg: "no argument file provided",
			err: errors.New("no argument file provided"),
		})
	}

	b, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		exit(response{
			Msg: "could not read configuration",
			err: err,
		})
	}

	var build builder
	err = json.Unmarshal(b, &build)
	if err != nil {
		exit(response{
			Msg: "configuration file is not valid JSON",
			err: err,
		})
	}

	exit(build.run())
}

func exit(resp response) {
	if resp.err != nil {
		resp.Failed = true
		resp.Err = resp.err.Error()
	}
	b, err := json.Marshal(resp)
	if err != nil {
		b, _ = json.Marshal(response{
			Msg:    "invalid response",
			Failed: true,
			Err:    err.Error(),
		})
	}
	fmt.Printf("%s\n", b)
	if resp.Failed {
		os.Exit(1)
	}
	os.Exit(0)
}

type response struct {
	Msg     string   `json:"msg"`
	Cmd     []string `json:"cmd,omitempty"`
	Changed bool     `json:"changed"`
	Failed  bool     `json:"failed"`
	Err     string   `json:"err,omitempty"`

	Stdout string `json:"stdout,omitempty"`
	Stderr string `json:"stderr,omitempty"`

	err error
}

type builder struct {
	GOROOT  string   `json:"goroot"`
	Pkg     string   `json:"pkg"`
	Dir     string   `json:"dir"`
	Timeout duration `json:"timeout"`
	Count   string   `json:"count"`
}

func (b builder) run() response {
	var resp response

	resp.Cmd = []string{filepath.Join(b.GOROOT, "bin", "go"), "test"}
	if b.Timeout != 0 {
		resp.Cmd = append(resp.Cmd, fmt.Sprintf("-timeout=%v", time.Duration(b.Timeout)))
	}
	if b.Count != "" {
		count, err := strconv.Atoi(b.Count)
		if err != nil {
			resp.Msg = "configuration file contains an invalid count"
			resp.err = err
			exit(resp)
		}
		resp.Cmd = append(resp.Cmd, fmt.Sprintf("-count=%d", count))
	}
	resp.Cmd = append(resp.Cmd, b.Pkg)

	cmd := exec.Command(resp.Cmd[0], resp.Cmd[1:]...)
	cmd.Dir = b.Dir
	var buf, bufErr bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &bufErr

	err := cmd.Run()
	resp.Stdout = buf.String()
	resp.Stderr = bufErr.String()
	switch {
	case strings.Contains(resp.Stdout, "\nFAIL\n"):
		resp.Msg = "failed"
		resp.Failed = true
	case err != nil:
		resp.Msg = fmt.Sprintf("go test: %v", err)
		resp.err = err
	default:
		resp.Msg = "passed"
		resp.Changed = !allCached(resp.Stdout)
	}
	return resp
}

func allCached(s string) bool {
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		line := sc.Bytes()
		if bytes.HasPrefix(line, []byte("ok")) && !bytes.HasSuffix(line, []byte("(cached)")) {
			return false
		}
	}
	return true
}

type duration time.Duration

func (d *duration) UnmarshalJSON(b []byte) error {
	s, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	t, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = duration(t)
	return nil
}
