# AzureRM MSI Auth Proxy

A lightweight HTTP proxy for forwarding Azure Managed Service Identity (MSI) authentication requests. This proxy is designed to run in environments where you need to securely forward Azure identity requests, such as in containerized or hybrid cloud scenarios.

## Features

- Forwards HTTP requests to the Azure Instance Metadata Service (IMDS) endpoint
- Injects required identity headers and API version
- Robust error handling and detailed logging
- Simple configuration via environment variables

## Requirements

- Go 1.20 or newer
- Azure environment with Managed Identity enabled

## Building

To build the project and output the binary as `azurerm-msi-auth-proxy` in the `../server/` directory:

```bash
go build -o ../server/azurerm-msi-auth-proxy main.go
```

## Usage

Set the following environment variables before running the proxy:

- `IDENTITY_ENDPOINT`: [REQUIRED] The Azure IMDS endpoint URL. It should be pre-existing in MSI enabled environment. (e.g., `http://localhost:12356/msi/token`)
- `IDENTITY_HEADER`: [REQUIRED] The value for the `x-identity-header` required by Azure. It should be pre-existing in MSI enabled environment.
- `ARM_MSI_API_VERSION`: [OPTIONAL] The API version to use for MSI requests (e.g., `2019-08-01`)
- `ARM_MSI_API_PROXY_PORT`: [OPTIONAL] The port on which the proxy should listen (e.g., `8080`)

### Example

```bash
export IDENTITY_ENDPOINT="http://localhost:12356/msi/token"
export IDENTITY_HEADER="<your-identity-header>"
export ARM_MSI_API_VERSION="2019-08-01"
export ARM_MSI_API_PROXY_PORT="42300"

./azurerm-msi-auth-proxy
```

## Logging

The proxy logs all incoming requests, outgoing requests, errors, and startup/shutdown events to standard output. Sensitive values (such as the identity header) are never logged.

## Error Handling

- The proxy will terminate with a fatal error if any required environment variable is missing or invalid.
- All HTTP forwarding and response errors are logged with details.
- Panics in the HTTP handler are recovered and logged.

## License

MIT License

## Author

- vermacodes
