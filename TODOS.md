# Todo:

### General
- Add event tracking for clickhouse
- Fix the frontend

### Features
- Add metrics and monitoring (Prometheus and Grafana)
- Security
  - Self-signed TLS as a proof of concept
- CI/CD
  - Setup tests (unit and integration)
  - Setup Github Actions to run the CI/CD pipeline 
  - Do that once docker compose actually runs lol
- Gateway features
  - Rate limiting
- Resilence
  - Exponential backoffs (to prevent spamming services)
  - Resource limits
  - Better input validation

### Later Features
- Deployment to K8 (Terraform/Pulumi) (maybe use both just to say I can)
- Hashicorp
  - Change Hashicorp vault to no longer be in dev mode → persistent secrets
  - Dynamic Secrets (change secrets each time you access it)


### More features
- Email validation (send an email when registering)
- Add database migration (allow a way to non-destructively update database tables without downing the container)
