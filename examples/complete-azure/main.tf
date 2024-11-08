terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
    azuread = {
      source  = "hashicorp/azuread"
      version = "~> 2.0"
    }
    zenml = {
        source = "zenml-io/zenml"
    }
  }
}

provider "zenml" {
  server_url = var.zenml_server_url
  api_key    = var.zenml_api_key
}

provider "azurerm" {
    features {
        resource_group {
            prevent_deletion_if_contains_resources = false
        }
    }
}

data "azurerm_client_config" "current" {}
data "azuread_client_config" "current" {}
# Get current subscription details
data "azurerm_subscription" "primary" {
  # This will get the subscription ID from the provider configuration
}

# Create resource group to hold all resources

resource "azurerm_resource_group" "resource_group" {
  name     = "zenml-${var.environment}"
  location = var.location
}

# Create storage account and blob storage container to store ZenML artifacts

resource "azurerm_storage_account" "artifacts" {
  name                     = "zenml${var.environment}"
  resource_group_name      = azurerm_resource_group.resource_group.name
  location                 = azurerm_resource_group.resource_group.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_storage_container" "artifact_container" {
  name                  = "zenml${var.environment}"
  storage_account_name  = azurerm_storage_account.artifacts.name
  container_access_type = "private"
}

# Create Azure Container Registry to store ZenML containers

resource "azurerm_container_registry" "containers" {
  name                = "zenml${var.environment}"
  resource_group_name = azurerm_resource_group.resource_group.name
  location            = azurerm_resource_group.resource_group.location
  sku                 = "Basic"
  admin_enabled       = true
}

# Create AzureML workspace and all related resources to run ZenML pipelines

resource "azurerm_application_insights" "application_insights" {
  name                = "zenml-${var.environment}"
  resource_group_name = azurerm_resource_group.resource_group.name
  location            = azurerm_resource_group.resource_group.location
  application_type    = "web"
}

resource "azurerm_key_vault" "key_vault" {
  name                = "zenml-${var.environment}"
  resource_group_name = azurerm_resource_group.resource_group.name
  location            = azurerm_resource_group.resource_group.location
  tenant_id           = data.azurerm_client_config.current.tenant_id
  sku_name            = "standard"
}

resource "azurerm_machine_learning_workspace" "azureml_workspace" {
  name                    = "zenml-${var.environment}"
  resource_group_name     = azurerm_resource_group.resource_group.name
  location                = azurerm_resource_group.resource_group.location
  storage_account_id      = azurerm_storage_account.artifacts.id
  container_registry_id   = azurerm_container_registry.containers.id
  application_insights_id = azurerm_application_insights.application_insights.id
  key_vault_id            = azurerm_key_vault.key_vault.id
  public_network_access_enabled = true
  identity {
    type = "SystemAssigned"
  }
  sku_name = "Basic"
}

# Create Service Principal with required permissions and password

resource "azuread_application" "service_principal_app" {
  display_name = "zenml-${var.environment}"
  owners       = [data.azuread_client_config.current.object_id]
}

resource "azuread_service_principal" "service_principal" {
  client_id = azuread_application.service_principal_app.client_id
  owners       = [data.azuread_client_config.current.object_id]
}

resource "azuread_service_principal_password" "service_principal_password" {
  service_principal_id = azuread_service_principal.service_principal.object_id
}

# Assign roles to the Service Principal

resource "azurerm_role_assignment" "storage_blob_data_contributor_role" {
  scope                = azurerm_storage_account.artifacts.id
  role_definition_name = "Storage Blob Data Contributor"
  principal_id         = azuread_service_principal.service_principal.object_id
}

resource "azurerm_role_assignment" "acr_push_role" {
  scope                = azurerm_container_registry.containers.id
  role_definition_name = "AcrPush"
  principal_id         = azuread_service_principal.service_principal.object_id
}

resource "azurerm_role_assignment" "acr_pull_role" {
  scope                = azurerm_container_registry.containers.id
  role_definition_name = "AcrPull"
  principal_id         = azuread_service_principal.service_principal.object_id
}

resource "azurerm_role_assignment" "acr_contributor_role" {
  scope                = azurerm_container_registry.containers.id
  role_definition_name = "Contributor"
  principal_id         = azuread_service_principal.service_principal.object_id
}

resource "azurerm_role_assignment" "azureml_compute_operator" {
  scope                = azurerm_machine_learning_workspace.azureml_workspace.id
  role_definition_name = "AzureML Compute Operator"
  principal_id         = azuread_service_principal.service_principal.object_id
}

resource "azurerm_role_assignment" "azureml_data_scientist" {
  scope                = azurerm_machine_learning_workspace.azureml_workspace.id
  role_definition_name = "AzureML Data Scientist"
  principal_id         = azuread_service_principal.service_principal.object_id
}

# ZenML service Connector for Azure

resource "zenml_service_connector" "azure" {
  name           = "zenml-${var.environment}"
  type           = "azure"
  auth_method    = "service-principal"

  configuration = {
    subscription_id = "${data.azurerm_client_config.current.subscription_id}"
    resource_group = "${azurerm_resource_group.resource_group.name}"
    storage_account = "${azurerm_storage_account.artifacts.name}"
    tenant_id = "${data.azurerm_client_config.current.tenant_id}"
    client_id = "${azuread_application.service_principal_app.client_id}"
    client_secret = "${azuread_service_principal_password.service_principal_password.value}"
  }

  labels = {
    environment = var.environment
    managed_by  = "terraform"
  }
}

# Artifact Store Component
resource "zenml_stack_component" "artifact_store" {
  name      = "abc-${var.environment}"
  type      = "artifact_store"
  flavor    = "azure"

  configuration = {
    path = "az://${azurerm_storage_container.artifact_container.name}"
  }

  connector_id = zenml_service_connector.azure.id

  labels = {
    environment = var.environment
  }
}

# Container Registry Component
resource "zenml_stack_component" "container_registry" {
  name      = "acr-${var.environment}"
  type      = "container_registry"
  flavor    = "azure"

  configuration = {
    uri = azurerm_container_registry.containers.login_server
  }

  connector_id = zenml_service_connector.azure.id

  labels = {
    environment = var.environment
  }
}

# SageMaker Orchestrator
resource "zenml_stack_component" "orchestrator" {
  name      = "azureml-${var.environment}"
  type      = "orchestrator"
  flavor    = "azureml"

  configuration = {
    subscription_id = "${data.azurerm_client_config.current.subscription_id}"
    resource_group = "${azurerm_resource_group.resource_group.name}"
    workspace = "zenml-${var.environment}"
  }

  connector_id = zenml_service_connector.azure.id

  labels = {
    environment = var.environment
  }
}

# Complete Stack
resource "zenml_stack" "azure_stack" {
  name = "azure-${var.environment}"

  components = {
    artifact_store     = zenml_stack_component.artifact_store.id
    container_registry = zenml_stack_component.container_registry.id
    orchestrator      = zenml_stack_component.orchestrator.id
  }

  labels = {
    environment = var.environment
    managed_by  = "terraform"
  }
}