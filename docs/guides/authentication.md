---
page_title: "Authentication - ZenML Provider"
subcategory: ""
description: |-
  Authenticating with the ZenML Provider
---

# Authentication

The ZenML provider requires authentication to interact with your ZenML server. The provider uses API key authentication to obtain access tokens.

Configure the provider with your ZenML server URL and API key or API token.

```hcl
provider "zenml" {
  server_url = "https://your-zenml-server.com"
  api_key    = "your-api-key"
}
```

For OSS users, the `server_url` is basically the root URL of your ZenML server deployment.
For Pro users, the `server_url` is the the URL of your workspace, which can be found
in your dashboard:

![ZenML workspace URL](../../assets/workspace_url.png)

It should look like something like `https://1bfe8d94-zenml.cloudinfra.zenml.io`.

You have two options to provide a token or key:

#### Option 1: Using `ZENML_API_KEY`

You can input the `ZENML_API_KEY` as follows: 

```hcl
provider "zenml" {
  server_url = "https://your-zenml-server.com"
  api_key    = "your-api-key"
}
```

You can also use environment variables:

```bash
export ZENML_SERVER_URL="https://your-zenml-server.com"
export ZENML_API_KEY="your-api-key"
```

To generate a `ZENML_API_KEY`, follow these steps:

1. Install ZenML:
```bash
pip install zenml
```

2. Login to your ZenML server:
```bash
zenml login --url <API_URL>
```

3. Create a service account and get the API key:
```bash
zenml service-account create <MYSERVICEACCOUNTNAME>
```

This command will print out the `ZENML_API_KEY` that you can use with this provider.

#### Option 2: Using `ZENML_API_TOKEN`

Alternatively, you can use an API token for authentication:

```hcl
provider "zenml" {
  server_url = "https://your-zenml-server.com"
  api_token  = "your-api-token"
}
```

You can also use environment variables:
```bash
export ZENML_SERVER_URL="https://your-zenml-server.com"
export ZENML_API_TOKEN="your-api-token"
```

## Troubleshooting

If you encounter authentication errors:
1. Verify your server URL is correct and accessible
2. Ensure your API key is valid and not expired
3. Check that your server URL doesn't have a trailing slash