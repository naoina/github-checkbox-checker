# GitHub checkbox checker

![7dbc2da](assets/7dbc2da.jpg)

## Requirements

- Go 1.7 or later

## Build

```bash
go get -u -v ./...
CGO_ENABLED=0 go build
```

## Preparation

### Get Personal access token

https://github.com/settings/tokens

Required scopes of OAuth2 is `Full control of private repositories`.

![8b07f86](assets/8b07f86.jpg)

### Install Webhook

- URL path of Payload URL must be `/payload`
- Content type must be `application.json`
- Required event is `Issues`

![bfe28c8](assets/bfe28c8.jpg)

## Run

```bash
GITHUB_WEBHOOK_SECRET=<secret> GITHUB_ACCESS_TOKEN=<access-token> ./github_checkbox_checker
```
