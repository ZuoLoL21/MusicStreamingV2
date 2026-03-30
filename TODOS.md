# Todo:

### General

### Features
- Add Kafka for async updates (update weight in bandit system)
- Security
  - Self-signed TLS as a proof of concept
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
- Add Oauth as a proof of concept