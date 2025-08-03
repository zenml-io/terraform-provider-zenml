# ZenML Project Example

This example demonstrates how to create a ZenML project using the Terraform provider.

## Authentication Options

You have two options for authentication:

### Option 1: API Key (Recommended)
Use a service account API key. This is recommended for long-term automation.

```bash
# Set environment variables
export ZENML_SERVER_URL="https://your-zenml-server.com"
export ZENML_API_KEY="your-api-key"
```

Or use the tfvars file:
```hcl
zenml_server_url = "https://your-zenml-server.com"
zenml_api_key    = "your-api-key"
```

### Option 2: API Token
Use a JWT token directly. Note that tokens expire and are not recommended for automation.

```bash
# Set environment variables
export ZENML_SERVER_URL="https://your-zenml-server.com"
export ZENML_API_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

Or use the tfvars file:
```hcl
zenml_server_url = "https://your-zenml-server.com"
zenml_api_token  = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## Usage

1. Copy the example tfvars file:
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

2. Edit `terraform.tfvars` with your actual values:
   - Update `zenml_server_url` with your ZenML server URL
   - Add either `zenml_api_key` OR `zenml_api_token` (not both)
   - Customize the project settings

3. Run Terraform:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

## Troubleshooting

- **401 Authentication Error**: 
  - If using JWT token, make sure it's not expired and you're using `zenml_api_token`
  - If using API key, make sure it's valid and you're using `zenml_api_key`
  - Don't use both `api_key` and `api_token` at the same time

- **JWT Token Format Error**: 
  - Make sure JWT tokens start with `eyJ` and are complete
  - Use `zenml_api_token` for JWT tokens, not `zenml_api_key`