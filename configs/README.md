# Configs

This directory stores public configuration examples, localization assets, and the default Casbin model.

- `config.example.yaml` is safe to commit and uses secure local defaults.
- Local private config should live in an untracked file such as `configs/config.yaml`.
- `APP_CONFIG_PATH` can point the CLI to another config file.
- `locales/` contains at least `zh-CN` and `en-US`.
- `rbac_model.conf` is the default Casbin model used by the runtime container.
