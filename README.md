# sev

[![CI](https://github.com/winebarrel/sev/actions/workflows/ci.yml/badge.svg)](https://github.com/winebarrel/sev/actions/workflows/ci.yml)

A tool that retrieves AWS Secrets Manager / Parameter Store values, sets them to environment variables, and executes commands.

## Usage

```
Usage: sev --config="~/.sev.toml" <profile> <command> ... [flags]

Arguments:
  <profile>        Profile name.
  <command> ...    Command and arguments.

Flags:
  -h, --help                         Show help.
      --config="~/.sev.toml"         Config file path ($SEV_CONFIG).
      --default-profile=STRING       Fallback profile name ($SEV_DEFAULT_PROFILE).
      --[no-]override-aws-profile    Use AWS_PROFILE in sev config (enabled by default).
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
AWS_PROFILE = "prof1" # By default, the AWS_PROFILE in the configuration file is used.
FOO = "secretsmanager://foo/bar"
ZOO = "secretsmanager://foo/zoo:TOKEN"
BAZ = "BAZBAZBAZ"

[another]
HOGE = "secretsmanager://foo/zoo:SECRET"
FOGA = "secretsmanager://foo/bar"
PIYO = "PIYOPIYOPIYO"
```

```sh
# Run `env` command with extra environment variables
$ sev default -- env
FOO=BAZ
ZOO=AAA
BAZ=BAZBAZBAZ

$ sev another -- env
HOGE=BBB
FUGA=BAZ
PIYO=PIYOPIYOPIYO
```

## Get values from Parameter Store

```toml
[default]
AWS_PROFILE = "prof1"
FOO = "parameterstore:///foo/bar"
ZOO = "parameterstore:///foo/zoo"
BAZ = "BAZBAZBAZ"
```
