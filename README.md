# rnp_checker

Telegram bot for checking RENAPER trámite status.

## Disclaimer

This project is created exclusively for demonstration and educational purposes.
It is an unofficial, community-driven project and is not affiliated with, endorsed by, or supported by RENAPER, the Argentine government, Telegram, or any other official entity.
The project is provided as-is, without any warranty, and the authors and contributors shall not be held liable for any direct, indirect, incidental, or consequential damages arising from its use.

This repository is distributed under the MIT License unless otherwise stated in the repository files.

## CI/CD

The repository includes GitHub Actions workflows for:

- `ci.yml` — runs formatting, vetting, and tests on push/PR.
- `deploy.yml` — builds a container image and deploys it to Google Cloud Run.

### Required GitHub variables

Set the following repository/environment variables in GitHub:

- `GCP_PROJECT_ID`
- `GCP_REGION` (optional, defaults to `us-central1`)
- `CLOUD_RUN_SERVICE_NAME` (optional, defaults to `rnp-checker`)
- `GCP_WORKLOAD_IDENTITY_PROVIDER`
- `GCP_SERVICE_ACCOUNT`

### Required GitHub secrets / Cloud Run secrets

- `telegram-token` secret (or equivalent) containing the Telegram bot token

### Runtime environment

Cloud Run should receive:

- `PORT=8080`
- `USE_WEBHOOK=true`
- `LANGUAGE=en`
- `PUBLIC_URL` (set by Cloud Run deployment output or manually)
