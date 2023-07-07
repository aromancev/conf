output "mongo_group_id" {
  description = "Group that has permissions for /opt/mongo"
  value       = local.mongo_group_id
}

output "mongo_user_id" {
  description = "User that has permissions for /opt/mongo"
  value       = local.mongo_user_id
}
