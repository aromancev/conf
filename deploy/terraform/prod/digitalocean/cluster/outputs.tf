output "mongo_user_id" {
  description = "User that has permissions for /opt/mongo/data"
  value       = local.mongo_user_id
}
