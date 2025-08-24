resource "kubernetes_secret_v1" "postgres_secret" {
  metadata {
    name      = "postgres-secrets"
    namespace = "db"
  }
  type = "Opaque"
  data = {
    POSTGRES_USER     = var.postgres_user
    POSTGRES_PASSWORD = var.postgres_password
    POSTGRES_DB       = var.postgres_db
    DATABASE_URL      = "postgres://${var.postgres_user}:${var.postgres_password}@postgres.db.svc.cluster.local:5432/${var.postgres_db}?sslmode=disable"
  }
  depends_on = [kubernetes_namespace.db]
}

resource "kubernetes_service_v1" "postgres" {
  metadata {
    name      = "postgres"
    namespace = "db"
  }
  spec {
    selector = {
      app = "postgres"
    }
    port {
      name        = "pg"
      port        = 5432
      target_port = 5432
    }
  }
  depends_on = [kubernetes_namespace.db]
}

resource "kubernetes_service_v1" "redis" {
  metadata {
    name      = "redis"
    namespace = "db"
  }
  spec {
    selector = {
      app = "redis"
    }
    port {
      name        = "redis"
      port        = 6379
      target_port = 6379
    }
  }
  depends_on = [kubernetes_namespace.db]
}

resource "kubernetes_stateful_set_v1" "postgres" {
  metadata {
    name      = "postgres"
    namespace = "db"
  }
  spec {
    service_name = kubernetes_service_v1.postgres.metadata[0].name
    replicas     = 1

    selector {
      match_labels = { app = "postgres" }
    }

    template {
      metadata { labels = { app = "postgres" } }
      spec {
        termination_grace_period_seconds = 30

        container {
          name  = "postgres"
          image = var.postgres_image

          port { container_port = 5432 }

          env {
            name = "POSTGRES_USER"
            value_from {
              secret_key_ref {
                name = kubernetes_secret_v1.postgres_secret.metadata[0].name
                key  = "POSTGRES_USER"
              }
            }
          }
          env {
            name = "POSTGRES_PASSWORD"
            value_from {
              secret_key_ref {
                name = kubernetes_secret_v1.postgres_secret.metadata[0].name
                key  = "POSTGRES_PASSWORD"
              }
            }
          }
          env {
            name = "POSTGRES_DB"
            value_from {
              secret_key_ref {
                name = kubernetes_secret_v1.postgres_secret.metadata[0].name
                key  = "POSTGRES_DB"
              }
            }
          }

          volume_mount {
            name       = "pgdata"
            mount_path = "/var/lib/postgresql/data"
          }

          readiness_probe {
            exec { command = ["sh", "-c", "pg_isready -U $POSTGRES_USER -h 127.0.0.1 -p 5432"] }
            initial_delay_seconds = 10
            period_seconds        = 5
          }

          liveness_probe {
            tcp_socket { port = 5432 }
            initial_delay_seconds = 20
            period_seconds        = 10
          }
        }

        volume {
          name = "pgdata"
          host_path {
            path = "/var/lib/postgres-data"
            type = "DirectoryOrCreate"
          }
        }
      }
    }
  }
  depends_on = [kubernetes_service_v1.postgres, kubernetes_secret_v1.postgres_secret]
}

resource "kubernetes_stateful_set_v1" "redis" {
  metadata {
    name      = "redis"
    namespace = "db"
  }
  spec {
    service_name = kubernetes_service_v1.redis.metadata[0].name
    replicas     = 1

    selector {
      match_labels = { app = "redis" }
    }

    template {
      metadata { labels = { app = "redis" } }
      spec {
        termination_grace_period_seconds = 10

        container {
          name  = "redis"
          image = var.redis_image
          args  = ["--appendonly", "yes"]

          port { container_port = 6379 }

          liveness_probe {
            tcp_socket { port = 6379 }
            initial_delay_seconds = 10
            period_seconds        = 10
          }

          readiness_probe {
            tcp_socket { port = 6379 }
            initial_delay_seconds = 5
            period_seconds        = 5
          }

          volume_mount {
            name       = "redisdata"
            mount_path = "/data"
          }
        }

        volume {
          name = "redisdata"
          host_path {
            path = "/var/lib/redis-data"
            type = "DirectoryOrCreate"
          }
        }
      }
    }
  }
  depends_on = [kubernetes_service_v1.redis]
}

resource "kubernetes_job_v1" "goose_migrate" {
  metadata {
    name      = "goose-migrate"
    namespace = "db"
    labels    = { job = "goose-migrate" }
  }

  spec {
    backoff_limit              = 5
    ttl_seconds_after_finished = 600

    template {
      metadata { labels = { job = "goose-migrate" } }

      spec {
        restart_policy = "OnFailure"

        init_container {
          name  = "wait-for-postgres"
          image = var.postgres_image
          command = [
            "sh", "-c",
            "until pg_isready -h postgres.db.svc.cluster.local -p 5432; do echo 'waiting for postgres'; sleep 2; done"
          ]
        }

        container {
          name  = "goose"
          image = var.goose_image

          env {
            name = "DATABASE_URL"
            value_from {
              secret_key_ref {
                name = kubernetes_secret_v1.postgres_secret.metadata[0].name
                key  = "DATABASE_URL"
              }
            }
          }

          command = ["sh", "-c", var.goose_command]
        }
      }
    }
  }

  depends_on = [
    kubernetes_stateful_set_v1.postgres,
  ]
}

resource "kubernetes_service_v1" "bot" {
  metadata {
    name      = "bot"
    namespace = "bot"
  }
  spec {
    selector = { app = "bot" }
    port {
      name        = "http"
      port        = var.bot_service_port
      target_port = var.bot_container_port
    }
  }
  depends_on = [kubernetes_namespace.bot]
}

resource "kubernetes_deployment_v1" "bot" {
  metadata {
    name      = "bot"
    namespace = "bot"
    labels    = { app = "bot" }
  }

  spec {
    replicas = 1

    selector {
      match_labels = { app = "bot" }
    }

    template {
      metadata { labels = { app = "bot" } }
      spec {
        container {
          name  = "bot"
          image = var.bot_image

          port { container_port = var.bot_container_port }

          env {
            name  = "BOT_TOKEN"
            value = var.bot_token
          }
          env {
            name  = "POSTGRES_URL"
            value = "postgres://${var.postgres_user}:${var.postgres_password}@${kubernetes_service_v1.postgres.metadata[0].name}.db.svc.cluster.local:5432/${var.postgres_db}?sslmode=disable"
          }
          env {
            name  = "REDIS_URL"
            value = "${kubernetes_service_v1.redis.metadata[0].name}.db.svc.cluster.local:6379"
          }
          env {
            name  = "REDIS_DB_ID"
            value = var.bot_redis_db_id
          }
          env {
            name  = "PORT"
            value = var.bot_service_port
          }
          env {
            name  = "RATE_LIMIT_TIME"
            value = var.bot_rate_limit_time
          }
          env {
            name  = "RATE_LIMIT_CHECK_INTERVAL"
            value = var.bot_rate_limit_check_interval
          }
          env {
            name  = "RATE_LIMIT_MAX_MESSAGES"
            value = var.bot_rate_limit_max_messages
          }

          liveness_probe {
            tcp_socket { port = var.bot_container_port }
            initial_delay_seconds = 10
            period_seconds        = 10
          }
          readiness_probe {
            tcp_socket { port = var.bot_container_port }
            initial_delay_seconds = 5
            period_seconds        = 5
          }
        }
      }
    }
  }

  depends_on = [
    kubernetes_service_v1.bot,
    kubernetes_service_v1.postgres,
    kubernetes_service_v1.redis
  ]
}

resource "kubernetes_config_map_v1" "fetcher_envs" {
  metadata {
    name      = "fetcher-envs"
    namespace = "fetcher"
  }
  data       = local.fetcher_envs
  depends_on = [kubernetes_namespace.fetcher]
}

resource "kubernetes_service_v1" "fetcher_headless" {
  metadata {
    name      = "fetcher"
    namespace = "fetcher"
  }
  spec {
    cluster_ip = "None"
    selector   = { app = "fetcher" }
    port {
      name        = "http"
      port        = var.fetcher_service_port
      target_port = var.fetcher_container_port
    }
  }
  depends_on = [kubernetes_namespace.fetcher]
}

resource "kubernetes_stateful_set_v1" "fetcher" {
  metadata {
    name      = "fetcher"
    namespace = "fetcher"
  }

  spec {
    service_name = kubernetes_service_v1.fetcher_headless.metadata[0].name
    replicas     = var.fetcher_replicas

    selector {
      match_labels = { app = "fetcher" }
    }

    template {
      metadata { labels = { app = "fetcher" } }
      spec {
        termination_grace_period_seconds = 10

        init_container {
          name  = "prepare-session"
          image = "busybox:1.36"
          command = [
            "sh", "-c",
            "set -e; mkdir -p /data; [ -f /data/session.json ] || touch /data/session.json; chmod 666 /data/session.json"
          ]
          volume_mount {
            name       = "fetcher-data"
            mount_path = "/data"
          }
        }

        container {
          name  = "fetcher"
          image = var.fetcher_image

          stdin      = true
          stdin_once = true
          tty        = false

          port { container_port = var.fetcher_container_port }

          env {
            name = "POD_NAME"
            value_from {
              field_ref {
                field_path = "metadata.name"
              }
            }
          }
          env {
            name  = "IP"
            value = "$(POD_NAME).fetcher.fetcher.svc.cluster.local"
          }
          env {
            name  = "PORT"
            value = var.fetcher_service_port
          }
          env {
            name  = "API"
            value = "bot.bot.svc.cluster.local:${var.bot_service_port}"
          }
          env {
            name  = "BOT_USERNAME"
            value = var.bot_username
          }
          env {
            name  = "REDIS_URL"
            value = "${kubernetes_service_v1.redis.metadata[0].name}.db.svc.cluster.local:6379"
          }

          volume_mount {
            name          = "envfile"
            mount_path    = "/app/.env"
            sub_path_expr = "$(POD_NAME).env"
          }

          volume_mount {
            name       = "fetcher-data"
            mount_path = "/app/session.json"
            sub_path   = "session.json"
          }

          liveness_probe {
            tcp_socket { port = var.fetcher_container_port }
            initial_delay_seconds = 10
            period_seconds        = 10
          }
          readiness_probe {
            tcp_socket { port = var.fetcher_container_port }
            initial_delay_seconds = 5
            period_seconds        = 5
          }
        }

        volume {
          name = "envfile"
          config_map {
            name = kubernetes_config_map_v1.fetcher_envs.metadata[0].name
          }
        }
      }
    }

    volume_claim_template {
      metadata { name = "fetcher-data" }
      spec {
        access_modes = ["ReadWriteOnce"]
        resources { requests = { storage = "1Gi" } }
      }
    }
  }

  depends_on = [
    kubernetes_config_map_v1.fetcher_envs,
    kubernetes_service_v1.fetcher_headless,
    kubernetes_service_v1.redis
  ]
}
