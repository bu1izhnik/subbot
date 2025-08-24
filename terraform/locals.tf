locals {
  fetcher_envs = {
    for f in fileset(var.fetcher_env_dir, "*.env") :
    f => file("${var.fetcher_env_dir}/${f}")
  }
}