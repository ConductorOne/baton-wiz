![Baton Logo](./docs/images/baton-logo.png)

# `baton-wiz` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-wiz.svg)](https://pkg.go.dev/github.com/conductorone/baton-wiz) ![main ci](https://github.com/conductorone/baton-wiz/actions/workflows/main.yaml/badge.svg)

`baton-wiz` is a connector for built using the [Baton SDK](https://github.com/conductorone/baton-sdk).

Check out [Baton](https://github.com/conductorone/baton) to learn more the project in general.

# Getting Started

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-wiz
baton-wiz
baton resources
```

## docker

```
docker run --rm -v $(pwd):/out -e BATON_WIZ_CLIENT_ID=clientID -e BATON_WIZ_CLIENT_SECRET=clientSecret BATON_AUTH_URL=auth_url -e BATON_ENDPOINT_URL=auth_url -e BATON_AUDIENCE=audience -e BATON_RESOURCE_IDS=resourecID1,resourceID2  ghcr.io/conductorone/baton-wiz:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-wiz/cmd/baton-wiz@main

BATON_WIZ_CLIENT_ID=clientID \
BATON_WIZ_CLIENT_SECRET=clientSecret' \
BATON_ENDPOINT_URL=https://api.<region>.app.wiz.io/graphql \
BATON_AUTH_URL=https://auth.app.wiz.io/oauth/token \
BATON_AUDIENCE=wiz-api \
BATON_RESOURCE_IDS=resourceID1,resourceID2 baton-wiz

baton resources
```

# Data Model

`baton-wiz` will pull down information about the following resources:
- Users
- Wiz Resources

# Contributing, Support and Issues

We started Baton because we were tired of taking screenshots and manually
building spreadsheets. We welcome contributions, and ideas, no matter how
small&mdash;our goal is to make identity and permissions sprawl less painful for
everyone. If you have questions, problems, or ideas: Please open a GitHub Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

# `baton-wiz` Command Line Usage

```
baton-wiz

Usage:
  baton-wiz [flags]
  baton-wiz [command]

Available Commands:
  capabilities       Get connector capabilities
  completion         Generate the autocompletion script for the specified shell
  help               Help about any command

Flags:
      --client-id string             The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string         The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
      --wiz-client-id string         The Wiz client ID ($BATON_WIZ_CLIENT_ID)
      --wiz-client-secret string     The Wiz client secret ($BATON_WIZ_CLIENT_SECRET)
      --auth-url string              The Token URL for authentication with the Wiz API ($BATON_AUTH_URL)
      --endpoint-url                 The Wiz GraphQL API endpoint URL ($BATON_ENDPOINT_URL)
      --audience                     The Wiz audience ($BATON_AUDIENCE)
      --resource-ids                 The Wiz resource IDs to sync  ($BATON_RESOURCE_IDS)
  -f, --file string                  The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                         help for baton-wiz
      --log-format string            The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string             The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
  -p, --provisioning                 If this connector supports provisioning, this must be set in order for provisioning actions to be enabled ($BATON_PROVISIONING)
      --ticketing                    This must be set to enable ticketing support ($BATON_TICKETING)
  -v, --version                      version for baton-wiz

Use "baton-wiz [command] --help" for more information about a command.
```
