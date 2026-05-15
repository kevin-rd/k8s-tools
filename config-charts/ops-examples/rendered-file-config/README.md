# rendered-file-config

Optional CD/OPS examples for wiring a config chart to `biz-charts/generic`.

These files are not required in a business repository. They show how the OPS/CD
side can provide the app chart values and the Secret consumed by
`generic.fileConfig.render`.

## Files

- `generic-values.yaml`: example values for `biz-charts/generic`.
- `secret.example.yaml`: plain Kubernetes Secret example for clusters without
  ESO.
- `externalsecret.example.yaml`: External Secrets Operator example that creates
  the same Secret name used by `generic-values.yaml`.

The Secret can also be created by SOPS or SealedSecret, as long as it uses the
same name and keys expected by the config template.
