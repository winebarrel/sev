# sev

[![CI](https://github.com/winebarrel/sev/actions/workflows/ci.yml/badge.svg)](https://github.com/winebarrel/sev/actions/workflows/ci.yml)

A tool that retrieves AWS Secrets Manager secret values, sets them to environment variables, and executes commands.

## Usage

```
Usage: sev --config="~/.sev.toml" <profile> <command> ... [flags]

Arguments:
  <profile>        Profile name.
  <command> ...    Command and arguments.

Flags:
  -h, --help                    Show help.
      --config="~/.sev.toml"    Config file path ($SEV_CONFIG).
      --version
```

## Example

```sh
$ aws secretsmanager get-secret-value --secret-id foo/bar
{
  ...
  "SecretString": "BAZ",
  ...

$ aws secretsmanager get-secret-value --secret-id foo/zoo # JSON secret
{
  ...
  "SecretString": "{\"TOKEN\":\"AAA\",\"SECRET\":\"BBB\"}",
  ...
```

```sh
$ cat ~/.sev.toml
[default]
BAR = "foo/bar"
ZOO = "foo/zoo:TOKEN"

[another]
HOGE = "foo/zoo:SECRET"
FOGA = "foo/bar"
```

```sh
# Run `env` command with extra environment variables
$ sev default -- env
FOO=BAZ
ZOO=AAA

$ sev another -- env
HOGE=BBB
FUGA=BAZ
```
