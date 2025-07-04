output "workspace_id" {
  description = "ID of the created workspace"
  value       = zenml_workspace.alpha.id
}

output "workspace_url" {
  description = "URL of the created workspace"
  value       = zenml_workspace.alpha.url
}

output "workspace_status" {
  description = "Status of the created workspace"
  value       = zenml_workspace.alpha.status
}

output "team_ids" {
  description = "IDs of the created teams"
  value = {
    developers   = zenml_team.alpha_dev.id
    ml_engineers = zenml_team.alpha_ml.id
    ml_ops       = zenml_team.alpha_ops.id
  }
}

output "project_id" {
  description = "ID of the created project"
  value       = zenml_project.recommendation.id
}

output "dev_stack_id" {
  description = "ID of the development stack"
  value       = zenml_stack.dev.id
}

output "prod_stack_id" {
  description = "ID of the production stack"
  value       = zenml_stack.prod.id
}

output "service_connector_id" {
  description = "ID of the AWS service connector"
  value       = zenml_service_connector.aws.id
}

output "stack_components" {
  description = "IDs of the created stack components"
  value = {
    artifact_store      = zenml_stack_component.artifact_store.id
    container_registry  = zenml_stack_component.container_registry.id
    orchestrator_dev    = zenml_stack_component.orchestrator_dev.id
    orchestrator_prod   = zenml_stack_component.orchestrator_prod.id
    model_registry      = zenml_stack_component.model_registry.id
  }
}

output "role_assignments" {
  description = "Summary of role assignments"
  value = {
    workspace_assignments = {
      for k, v in zenml_workspace_role_assignment.teams : k => {
        team_id = v.team_id
        role    = v.role
      }
    }
    project_assignments = {
      for k, v in zenml_project_role_assignment.teams : k => {
        team_id = v.team_id
        role    = v.role
      }
    }
    stack_assignments = {
      for k, v in zenml_stack_role_assignment.permissions : k => {
        stack_id = v.stack_id
        team_id  = v.team_id
        role     = v.role
      }
    }
  }
}