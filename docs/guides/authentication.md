---
page_title: "Authentication - ZenML Provider"
subcategory: ""
description: |-
  Authenticating with the ZenML Provider
---

# Authentication

The ZenML provider requires authentication to interact with your ZenML server. This guide explains how to set up authentication for the provider.

## Configuration

The provider can be configured with the following environment variables:

- `ZENML_SERVER_URL`: The URL of your ZenML server
- `ZENML_API_KEY`: Your ZenML API key

Alternatively, you can provide these credentials in the provider configuration:

```hcl
provider "zenml" {
  server_url = "https://your-zenml-server.com"
  api_key    = "your-api-key"
}
```

!> **Warning:** Hard-coding credentials into your Terraform configuration is not recommended. Use environment variables or other secure methods to provide credentials.

## Obtaining Credentials

1. **Server URL**: This is the URL where your ZenML server is hosted.

2. **API Key**: You can generate an API key from the ZenML UI or CLI:
   - UI: Navigate to your user settings and create a new API key
   - CLI: Use the command `zenml api-key create --name="terraform" --description="For Terraform"`

## Best Practices

- Use environment variables or a secure secret management system to handle credentials.
- Rotate your API keys regularly.
- Use separate API keys for different environments (development, staging, production).

For more information on ZenML authentication, refer to the [ZenML documentation](https://docs.zenml.io/user-guide/advanced-guide/environment-management/connect-to-zenml).