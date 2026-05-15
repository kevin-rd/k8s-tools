# config-charts

CD/OPS-side standards and examples for publishing business application config
charts.

Chinese version: [README.zh-CN.md](README.zh-CN.md).

## Directory Layout

```text
config-charts/
  standards/
    config-chart/
  business-repo-example/
    configs/
      config-chart/
        README.md
        files/
          config.yaml
        values/
          values-devnet.yaml
          values-testnet.yaml
  ops-examples/
    rendered-file-config/
  docs/
    business-ci-prompt.md
    business-ci-prompt.zh-CN.md
```

- `standards/config-chart`: the shared config chart template that CI should copy.
- `business-repo-example`: the expected config source layout in a business
  repository.
- `ops-examples`: optional CD/OPS wiring examples, such as app chart values and
  Secret providers.
- `docs`: reusable prompts and longer implementation notes.

## Business Repository Contract

Business repositories should keep only config sources:

```text
configs/
  config-chart/
    README.md
    files/
      config.yaml
    values/
      values-devnet.yaml
      values-testnet.yaml
      values-mainnet.yaml
```

- `configs/config-chart/files/*`: config file templates that become ConfigMap
  entries.
- `configs/config-chart/values/values-<env>.yaml`: non-sensitive environment
  values used while rendering the config chart.
- Sensitive values must not be committed. Use gomplate placeholders such as
  `[[ env.Env.POSTGRES_PASSWORD ]]`.
- Do not add a top-level business `values.yaml`. The current convention is one
  `values/values-<env>.yaml` file per environment.

Keep business config values under `appConfig` to avoid collisions with standard
chart control values:

```yaml
appConfig:
  server:
    port: 8080
```

## Standard Chart Contract

`standards/config-chart` is a template, not a complete business chart. CI should
copy it into a temporary directory, inject business config files and all
environment values files, then validate and package the assembled chart.

The standard chart is published to:

```text
oci://ghcr.io/<owner>/biz-charts/config-chart
```

The standard chart package contains all environments. CD decides which
`values/values-<env>.yaml` file to use when installing the config chart.

Do not run `helm lint` directly against `standards/config-chart` before adding
`configs/*`; the template intentionally does not contain business config files.

## Naming

Use separate release names for the application and its config chart:

```text
app release:    demo-app
config release: demo-app-config
```

The config chart name should normally be `${APP_NAME}-config`, for example
`demo-app-config`.

The standard chart defaults its ConfigMap name from the effective release name:

- if the name already ends with `-config`, it is used as-is;
- otherwise `-config` is appended;
- `configMap.name` can override the generated name.

With the names above, the generated ConfigMap name is `demo-app-config`, and the
application chart can reference it through:

```yaml
fileConfig:
  enabled: true
  existingName: demo-app-config
  mountPath: /app/config
```

## Template Syntax

Config files are rendered in two phases:

1. Helm renders the assembled config chart and creates a ConfigMap.
2. The application pod runs `gomplate` in an initContainer when
   `generic.fileConfig.render.enabled=true`.

Use Helm `{{ }}` only for non-sensitive config values:

```yaml
server:
  port: {{ .Values.appConfig.server.port }}
```

Use gomplate `[[ ]]` for runtime Secret values:

```yaml
database:
  password: '[[ env.Env.POSTGRES_PASSWORD ]]'
```

The `[[ ]]` delimiter avoids conflicts with Helm templates.

## CI Assembly Flow

On a business application tag, CI should:

1. Build and push the application image.
2. Fetch this repository at a pinned tag or commit.
3. Copy `config-charts/standards/config-chart` into a temporary build directory.
4. Copy `configs/config-chart/files/*` into the chart `configs/` directory.
5. Copy `configs/config-chart/values/*` into the chart `values/` directory.
6. Set `Chart.yaml` fields:
   - `name: ${APP_NAME}-config`
   - `version: ${APP_VERSION}`
   - `appVersion: ${APP_TAG}`
7. Validate the assembled chart with each supported environment values file.
8. Package and push one config chart containing all environment values files.

Example shell outline:

```bash
APP_NAME=demo-app
APP_TAG=v1.2.3
APP_VERSION="${APP_TAG#v}"
CONFIG_CHART_NAME="${APP_NAME}-config"
BUSINESS_CONFIG_DIR=configs/config-chart
TEMPLATE_REF=v0.1.0
export APP_NAME APP_TAG APP_VERSION CONFIG_CHART_NAME TEMPLATE_REF

mkdir -p .build

git clone --depth 1 --branch "${TEMPLATE_REF}" \
  git@github.com:kevin-rd/k8s-tools.git .build/k8s-tools

cp -R .build/k8s-tools/config-charts/standards/config-chart \
  ".build/${CONFIG_CHART_NAME}"

mkdir -p ".build/${CONFIG_CHART_NAME}/configs" \
  ".build/${CONFIG_CHART_NAME}/values"
rsync -a "${BUSINESS_CONFIG_DIR}/files/" ".build/${CONFIG_CHART_NAME}/configs/"
rsync -a "${BUSINESS_CONFIG_DIR}/values/" ".build/${CONFIG_CHART_NAME}/values/"

yq -i '
  .name = strenv(CONFIG_CHART_NAME) |
  .version = strenv(APP_VERSION) |
  .appVersion = strenv(APP_TAG)
' ".build/${CONFIG_CHART_NAME}/Chart.yaml"

for values_file in ".build/${CONFIG_CHART_NAME}"/values/values-*.yaml; do
  helm lint ".build/${CONFIG_CHART_NAME}" -f "${values_file}"
  helm template "${CONFIG_CHART_NAME}" ".build/${CONFIG_CHART_NAME}" \
    -f "${values_file}" >/dev/null
done

helm package ".build/${CONFIG_CHART_NAME}" --destination .build/charts
helm push ".build/charts/${CONFIG_CHART_NAME}-${APP_VERSION}.tgz" \
  oci://ghcr.io/example/helm-charts
```

Do not commit the generated chart back to the business repository. Treat it as a
CI artifact.

## CD Order

CD should install the config chart before the application chart.

When installing the config chart, CD must provide the selected environment
values file:

```bash
helm upgrade --install demo-app-config \
  oci://ghcr.io/example/helm-charts/demo-app-config \
  --version 1.2.3 \
  -f values/values-devnet.yaml \
  -n demo
```

If CD deploys directly from an OCI or tgz chart artifact, it must first fetch or
extract the selected `values/values-<env>.yaml` file and then pass it to Helm
with `-f`.

Then install the application chart:

```bash
helm upgrade --install demo-app \
  oci://ghcr.io/example/helm-charts/generic \
  -f app-values.yaml \
  -n demo
```

The Secret used by `fileConfig.render.secret.existingName` can come from a plain
Kubernetes Secret, SOPS, SealedSecret, or External Secrets Operator. It only
needs to exist before the application pod starts.

## AI Prompt

Use [docs/business-ci-prompt.md](docs/business-ci-prompt.md) when adding config
chart CI to a business repository.
