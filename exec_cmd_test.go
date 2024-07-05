package sev

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_execCmd_OK(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	cmd := []string{"/usr/bin/env"}
	env := map[string]string{
		"FOO": "BAR",
		"ZOO": "BAZ",
	}

	bufout := &bytes.Buffer{}
	buferr := &bytes.Buffer{}
	_stdout = bufout
	_stderr = buferr

	defer func() {
		_stdout = os.Stdout
		_stderr = os.Stderr
	}()

	err := execCmd(cmd, env)
	require.NoError(err)
	assert.Contains(bufout.String(), "FOO=BAR\n")
	assert.Contains(bufout.String(), "ZOO=BAZ\n")
	assert.Empty(buferr.String())
}

func Test_execCmd_WithArgs(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	cmd := []string{"/bin/sh", "-c", "echo $FOO $ZOO"}
	env := map[string]string{
		"FOO": "BAR",
		"ZOO": "BAZ",
	}

	bufout := &bytes.Buffer{}
	buferr := &bytes.Buffer{}
	_stdout = bufout
	_stderr = buferr

	defer func() {
		_stdout = os.Stdout
		_stderr = os.Stderr
	}()

	err := execCmd(cmd, env)
	require.NoError(err)
	assert.Equal("BAR BAZ\n", bufout.String())
	assert.Empty(buferr.String())
}

func Test_execCmd_Err(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	cmd := []string{"/bin/ls", "/not/exists"}
	env := map[string]string{
		"FOO": "BAR",
		"ZOO": "BAZ",
	}

	bufout := &bytes.Buffer{}
	buferr := &bytes.Buffer{}
	_stdout = bufout
	_stderr = buferr

	defer func() {
		_stdout = os.Stdout
		_stderr = os.Stderr
	}()

	err := execCmd(cmd, env)
	require.Error(err)
	assert.Empty(bufout.String())
	assert.NotEmpty(buferr.String())
}
