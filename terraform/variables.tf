variable "cluster_name" {
  type    = string
  default = "subbot"
}


# Images
variable "postgres_image" {
  type    = string
  default = "postgres:16"
}
variable "redis_image" {
  type    = string
  default = "redis:7"
}
variable "bot_image" {
  type    = string
  default = "bu1izhnik/subbot:latest"
}
variable "fetcher_image" {
  type    = string
  default = "bu1izhnik/subfetcher:latest"
}
variable "goose_image" {
  type    = string
  default = "bu1izhnik/subgoose:latest"
}

variable "goose_command" {
  type    = string
  default = "goose -dir /migrations postgres \"$DATABASE_URL\" up"
}

variable "postgres_user" {
  type    = string
  default = "subbot"
}
variable "postgres_password" {
  type    = string
  default = "1234"
}
variable "postgres_db" {
  type    = string
  default = "subbot"
}

variable "bot_container_port" {
  type    = number
  default = 8080
}
variable "bot_service_port" {
  type    = number
  default = 8080
}
variable "bot_redis_db_id" {
  type    = number
  default = 0
}
variable "bot_rate_limit_time" {
  type    = number
  default = 300
}
variable "bot_rate_limit_check_interval" {
  type    = number
  default = 60
}
variable "bot_rate_limit_max_messages" {
  type    = number
  default = 20
}
variable "bot_token" {
  type        = string
  description = "telegram api bot token"
}
variable "bot_username" {
  type        = string
  default     = "stonesubbot"
  description = "main bot username"
}
variable "bot_liveness_path" {
  type    = string
  default = "/healthz"
}
variable "bot_readiness_path" {
  type    = string
  default = "/readyz"
}

variable "fetcher_container_port" {
  type    = number
  default = 8081
}
variable "fetcher_service_port" {
  type    = number
  default = 8081
}
variable "fetcher_replicas" {
  type    = number
  default = 1
}

variable "fetcher_env_dir" {
  type        = string
  default     = "envs"
  description = "Path to a folder with *.env files for fetcher replicas."

  validation {
    condition     = length(fileset(var.fetcher_env_dir, "*.env")) > 0
    error_message = "fetcher_env_dir must contain at least one .env file."
  }
}
