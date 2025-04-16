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
  config             Get the connector config schema
  help               Help about any command

Flags:
      --audience string                                  The audience used to authenticate with Wiz ($BATON_AUDIENCE) (default "wiz-api")
      --auth-url string                                  required: The auth url used to authenticate with Wiz ($BATON_AUTH_URL)
      --client-id string                                 The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string                             The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
      --endpoint-url string                              required: The endpoint url used to authenticate with Wiz ($BATON_ENDPOINT_URL)
      --external-resource-c1z string                     The path to the c1z file to sync external baton resources with ($BATON_EXTERNAL_RESOURCE_C1Z)
      --external-resource-entitlement-id-filter string   The entitlement that external users, groups must have access to sync external baton resources ($BATON_EXTERNAL_RESOURCE_ENTITLEMENT_ID_FILTER)
      --external-sync-mode                               Enable external sync mode ($BATON_EXTERNAL_SYNC_MODE)
  -f, --file string                                      The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                                             help for baton-wiz
      --log-format string                                The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string                                 The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
      --otel-collector-endpoint string                   The endpoint of the OpenTelemetry collector to send observability data to (used for both tracing and logging if specific endpoints are not provided) ($BATON_OTEL_COLLECTOR_ENDPOINT)
      --project-id string                                Scope the resource graph query to a specific project. Required if service account does not have access to all projects. ($BATON_PROJECT_ID)
  -p, --provisioning                                     This must be set in order for provisioning actions to be enabled ($BATON_PROVISIONING)
      --resource-ids strings                             The resource ids to sync ($BATON_RESOURCE_IDS)
      --skip-full-sync                                   This must be set to skip a full sync ($BATON_SKIP_FULL_SYNC)
      --sync-identities                                  Enable if wiz identities should be synced ($BATON_SYNC_IDENTITIES)
      --sync-service-accounts                            Enable if wiz service accounts should be synced ($BATON_SYNC_SERVICE_ACCOUNTS)
      --tags string                                      The tags on resources to sync ($BATON_TAGS)
      --ticketing                                        This must be set to enable ticketing support ($BATON_TICKETING)
  -v, --version                                          version for baton-wiz
      --wiz-client-id string                             required: The client ID used to authenticate with Wiz ($BATON_WIZ_CLIENT_ID)
      --wiz-client-secret string                         required: The client secret used to authenticate with Wiz ($BATON_WIZ_CLIENT_SECRET)
      --wiz-resource-types strings                       The wiz resource-types to sync ($BATON_WIZ_RESOURCE_TYPES)

Use "baton-wiz [command] --help" for more information about a command.
