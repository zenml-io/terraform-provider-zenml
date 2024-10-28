---
page_title: "Authentication - ZenML Provider"
subcategory: ""
description: |-
  Authenticating with the ZenML Provider
---

# Authentication

The ZenML provider requires authentication to interact with your ZenML server. The provider uses API key authentication to obtain access tokens.

## Configuration

The provider can be configured using environment variables:

* `ZENML_SERVER_URL` - (Required) The URL of your ZenML server
* `ZENML_API_KEY` - (Required) Your ZenML API key

Alternatively, you can provide these credentials directly in the provider configuration:

```hcl
provider "zenml" {
  server_url = "https://your-zenml-server.com"
  api_key    = "your-api-key"
}
```

!> **Warning:** Hard-coding credentials into your Terraform configuration is not recommended. Use environment variables or other secure methods to provide credentials.

## Authentication Process

The provider automatically handles the authentication process by:
1. Making a login request to `/api/v1/login` with your API key
2. Obtaining an access token
3. Using this token for subsequent API requests

The access token is automatically refreshed for each request to ensure continuous operation.

## Obtaining Credentials

1. **Server URL**: This is the URL where your ZenML server is hosted. For example: `https://your-zenml-server.com`

2. **API Key**: You can generate an API key from the ZenML UI or CLI:
   ```bash
   zenml api-key create --name="terraform" --description="For Terraform provider"
   ```

## Best Practices

* Store credentials using environment variables:
  ```bash
  export ZENML_SERVER_URL="https://your-zenml-server.com"
  export ZENML_API_KEY="your-api-key"
  ```
* Use different API keys for different environments
* Rotate API keys regularly
* Never commit API keys to version control

## Troubleshooting

If you encounter authentication errors:
1. Verify your server URL is correct and accessible
2. Ensure your API key is valid and not expired
3. Check that your server URL doesn't have a trailing slash