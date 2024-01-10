# security-review

Security Code Review using GitHub Actions 🤖.

- **SecretScanning**: Scan for sensitive information committed to source code.
- **CodeScanning**: Perform static analysis of source code to identify problem areas.
- **Comment**: Put review comments on PRs.

![image](image/pullrequest-review.png)

This tool allows you to shift-left security in your development environment💪

## Usage

Create workflow yaml (`.github/workflows/security-review.yaml`) on your repository.

```yaml
name: Security Code Review on PR
on:
  pull_request:
    branches:
      - main
    types: [opened, synchronize]
jobs:
  review:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: write # risken review needs this permission to create a comment on the PR

    steps:
      - uses: actions/checkout@v4
      - uses: ca-risken/security-review@v1
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
```

## Integrate RISKEN

[RISKEN](https://docs.security-hub.jp/) is a platform for collecting security issues; Findings detected by Actions can be linked to the RISKEN environment for issue management, alerting, information sharing to the team, and analysis results from the generated AI.

```yaml
- uses: ca-risken/security-review@v1
  with:
    github_token: ${{ secrets.GITHUB_TOKEN }}
    risken_console_url: ${{ env.RISKEN_CONSOLE_URL }}
    risken_api_endpoint: ${{ env.RISKEN_API_ENDPOINT }}
    risken_api_token: ${{ secrets.RISKEN_API_TOKEN }}
```

| Pameters | Description | Examples |
| ---- | ---- | ---- |
| `risken_console_url` | RISKEN Console URL | https://console.your-env.com |
| `risken_api_endpoint` | RISKEN API Endpoint | https://api.your-env.com |
| `risken_api_token` | RISKEN API Token | xxxxx |

## Other Options

```yaml
- uses: ca-risken/security-review@v1
  with:
    github_token: ${{ secrets.GITHUB_TOKEN }}
    options: '--error'
```

| Pameters | Description | Examples |
| ---- | ---- | ---- |
| `--error` | Exit 1 if there are finding (default: false) | --error |

## Test on local

### Generate ENV file

```shell
$ cp .env.sample .env
$ vi .env # fix your token
```

### Use Docker

```shell
$ make run
```

### See your comment

https://github.com/ca-risken/security-review/pull/1

## Push image

```shell
$ make push TAG=v1
```

