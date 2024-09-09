package sev_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winebarrel/sev"
)

func Test_loadEnfFrom_OK(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	tomlFile, _ := os.CreateTemp("", "")
	defer os.Remove(tomlFile.Name())
	tomlFile.WriteString(`[abc]
FOO = "BAR"
ZOO = "BAZ"
[DEF]
hoge = "fuga"
piyo = "hogera"
`)
	tomlFile.Sync()

	envAbc, err := sev.LoadEnvFrom(tomlFile.Name(), "abc", "")
	require.NoError(err)
	assert.Equal(map[string]string{"FOO": "BAR", "ZOO": "BAZ"}, envAbc)

	envDef, err := sev.LoadEnvFrom(tomlFile.Name(), "DEF", "")
	require.NoError(err)
	assert.Equal(map[string]string{"hoge": "fuga", "piyo": "hogera"}, envDef)
}

func Test_loadEnfFrom_Err_ConfigNotExists(t *testing.T) {
	assert := assert.New(t)
	_, err := sev.LoadEnvFrom("/not/exists", "abc", "")
	assert.ErrorContains(err, "no such file or directory")
}

func Test_loadEnfFrom_Err_ParseError(t *testing.T) {
	assert := assert.New(t)

	tomlFile, _ := os.CreateTemp("", "")
	defer os.Remove(tomlFile.Name())
	tomlFile.WriteString("xxx")
	tomlFile.Sync()

	_, err := sev.LoadEnvFrom(tomlFile.Name(), "abc", "")
	assert.ErrorContains(err, "expected key separator '='")
}

func Test_loadEnfFrom_ProfileNotFound(t *testing.T) {
	assert := assert.New(t)

	tomlFile, _ := os.CreateTemp("", "")
	defer os.Remove(tomlFile.Name())
	tomlFile.WriteString(`[abc]
FOO = "BAR"
ZOO = "BAZ"
`)
	tomlFile.Sync()

	_, err := sev.LoadEnvFrom(tomlFile.Name(), "DEF", "")
	assert.ErrorContains(err, "profile could not be found: DEF")
}

func Test_loadEnfFrom_Fallback(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	tomlFile, _ := os.CreateTemp("", "")
	defer os.Remove(tomlFile.Name())
	tomlFile.WriteString(`[xabcx]
FOO = "BAR"
ZOO = "BAZ"
[default]
FOO = "rab"
ZOO = "zab"
`)
	tomlFile.Sync()

	env, err := sev.LoadEnvFrom(tomlFile.Name(), "abc", "default")
	require.NoError(err)
	assert.Equal(map[string]string{"FOO": "rab", "ZOO": "zab"}, env)
}

func Test_loadEnfFrom_FallbackNotFound(t *testing.T) {
	assert := assert.New(t)

	tomlFile, _ := os.CreateTemp("", "")
	defer os.Remove(tomlFile.Name())
	tomlFile.WriteString(`[xabcx]
FOO = "BAR"
ZOO = "BAZ"
[xdefaultx]
FOO = "rab"
ZOO = "zab"
`)
	tomlFile.Sync()

	_, err := sev.LoadEnvFrom(tomlFile.Name(), "abc", "default")
	assert.ErrorContains(err, "fallback profile could not be found: default")
}
