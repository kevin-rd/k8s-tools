# config-charts

用于发布业务应用 config chart 的 CD/OPS 侧标准、模板和示例。

English version: [README.md](README.md).

## 目录结构

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

- `standards/config-chart`：CI 应复制使用的共享 config chart 模板。
- `business-repo-example`：业务代码库中推荐保留的配置源文件结构。
- `ops-examples`：可选的 CD/OPS 接入示例，例如 app chart values 和 Secret 来源。
- `docs`：可复用 prompt 和更长的接入说明。

## 业务代码库约定

业务代码库只保留配置源：

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

- `configs/config-chart/files/*`：最终进入 ConfigMap 的配置文件模板。
- `configs/config-chart/values/values-<env>.yaml`：config chart 渲染时使用的非敏感环境配置值。
- 敏感值不要提交到业务代码库，使用 gomplate 占位符，例如 `[[ env.Env.POSTGRES_PASSWORD ]]`。
- 不要添加总的业务 `values.yaml`。当前约定是每个环境一个 `values/values-<env>.yaml` 文件。

业务配置值建议统一放在 `appConfig` 下，避免和标准 chart 自身控制字段冲突：

```yaml
appConfig:
  server:
    port: 8080
```

## 标准 Chart 约定

`standards/config-chart` 是模板，不是一份完整的业务 chart。CI 应将它复制到临时目录，注入业务配置文件和所有环境 values 文件后，再校验并打包。

标准 chart 包会包含所有环境。CD 在安装 config chart 时决定使用哪个 `values/values-<env>.yaml` 文件。

不要在还没有加入 `configs/*` 前直接对 `standards/config-chart` 执行 `helm lint`；模板本身有意不包含业务配置文件。

## 命名约定

应用 chart 和 config chart 使用不同 release name：

```text
app release:    demo-app
config release: demo-app-config
```

config chart 的 chart name 通常也建议使用 `${APP_NAME}-config`，例如 `demo-app-config`。

标准 chart 会根据有效 release name 生成默认 ConfigMap 名：

- 如果 name 已经以 `-config` 结尾，直接使用；
- 如果 name 没有 `-config` 后缀，自动追加 `-config`；
- 也可以通过 `configMap.name` 显式覆盖。

按上面的命名，生成的 ConfigMap 名是 `demo-app-config`，应用 chart 可以这样引用：

```yaml
fileConfig:
  enabled: true
  existingName: demo-app-config
  mountPath: /app/config
```

## 模板语法

配置文件会经过两阶段渲染：

1. Helm 渲染组装后的 config chart，生成 ConfigMap。
2. 当 `generic.fileConfig.render.enabled=true` 时，应用 Pod 通过 initContainer 运行 `gomplate`，把 Secret 值渲染到最终配置文件。

Helm `{{ }}` 只用于非敏感配置：

```yaml
server:
  port: {{ .Values.appConfig.server.port }}
```

gomplate `[[ ]]` 用于运行时 Secret：

```yaml
database:
  password: '[[ env.Env.POSTGRES_PASSWORD ]]'
```

使用 `[[ ]]` 是为了避免和 Helm 的 Go template 语法冲突。

## CI 组装流程

业务应用打 tag 后，CI 建议按下面流程执行：

1. 构建并推送业务镜像。
2. 按固定 tag 或 commit 拉取本仓库。
3. 将 `config-charts/standards/config-chart` 复制到临时构建目录。
4. 将业务代码库中的 `configs/config-chart/files/*` 放入 chart 的 `configs/` 目录。
5. 将业务代码库中的 `configs/config-chart/values/*` 放入 chart 的 `values/` 目录。
6. 设置 `Chart.yaml`：
   - `name: ${APP_NAME}-config`
   - `version: ${APP_VERSION}`
   - `appVersion: ${APP_TAG}`
7. 用每个环境 values 文件校验组装后的 chart。
8. 打包并推送一个包含所有环境 values 文件的 config chart。

示例命令：

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

生成出来的 chart 不要提交回业务代码库，它只是 CI artifact。

## CD 发布顺序

CD 应先发布 config chart，再发布业务 app chart。

发布 config chart 时，CD 必须传入选中的环境 values 文件：

```bash
helm upgrade --install demo-app-config \
  oci://ghcr.io/example/helm-charts/demo-app-config \
  --version 1.2.3 \
  -f values/values-devnet.yaml \
  -n demo
```

如果 CD 直接从 OCI 或 tgz chart artifact 部署，需要先取得或解出其中的 `values/values-<env>.yaml` 文件，再通过 Helm `-f` 传入。

然后发布业务 app chart：

```bash
helm upgrade --install demo-app \
  oci://ghcr.io/example/helm-charts/generic \
  -f app-values.yaml \
  -n demo
```

`fileConfig.render.secret.existingName` 引用的 Secret 可以来自普通 Kubernetes Secret、SOPS、SealedSecret 或 External Secrets Operator。只要应用 Pod 启动前这个 Secret 已存在即可。

## AI Prompt

在业务代码库中接入 config chart 发布流程时，可以使用 [docs/business-ci-prompt.zh-CN.md](docs/business-ci-prompt.zh-CN.md)。
