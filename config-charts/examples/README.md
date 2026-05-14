# rendered-file-config

This example shows how a business repository can keep a complete config chart
that publishes non-sensitive config templates, while the application chart uses
`generic.fileConfig.render` to render sensitive values from a Kubernetes Secret
at pod startup.

## Files

- `config-chart/`: example config chart owned by the business repository.
- `generic-values.yaml`: values for `biz-charts/generic`.
- `secret.example.yaml`: plain Kubernetes Secret example for clusters without ESO.
- `externalsecret.example.yaml`: External Secrets Operator example that creates
  the same Secret name used by `generic-values.yaml`.

The config chart defaults its ConfigMap name from the effective release name:
if the name already ends with `-config`, it is used as-is; otherwise `-config`
is appended. `configMap.name` can override the generated name.

## Render locally

```bash
helm lint ./config-chart
helm template demo-app-config ./config-chart
helm template demo-app ../../../biz-charts/generic -f ./generic-values.yaml
```

## Install order

```bash
helm upgrade --install demo-app-config ./config-chart -n demo --create-namespace
kubectl apply -f secret.example.yaml -n demo
helm upgrade --install demo-app ../../../biz-charts/generic -f generic-values.yaml -n demo
```

The intended release names are `demo-app-config` for the config chart and
`demo-app` for the application chart. Both can reference the same rendered
ConfigMap name, `demo-app-config`.

The Secret can also be created by SOPS, SealedSecret, or External Secrets
Operator, as long as it uses the same name and keys expected by the config
template.
