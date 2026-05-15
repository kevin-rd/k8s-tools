# config-chart

这个目录是业务代码库中推荐保留的配置源目录。业务仓库只维护配置模板和各环境的非敏感 values，不维护完整 Helm chart 模板。

## 目录结构

```text
configs/config-chart/
  README.md
  files/
    config.yaml
  values/
    values-devnet.yaml
    values-testnet.yaml
```

- `files/*`：配置文件模板，最终会进入 config chart 生成的 ConfigMap。
- `values/values-<env>.yaml`：对应环境的非敏感配置值。
- 敏感值不要写入仓库，使用 gomplate 占位符，例如 `[[ env.Env.POSTGRES_PASSWORD ]]`。
- 不要添加总的 `values.yaml`。当前约定是每个环境一个 `values/values-<env>.yaml` 文件。

## 配置模板

`files/config.yaml` 中可以使用 Helm values 渲染非敏感配置：

```yaml
server:
  port: {{ .Values.appConfig.server.port }}
```

运行时 Secret 使用 gomplate `[[ ]]` 占位符：

```yaml
database:
  password: '[[ env.Env.POSTGRES_PASSWORD ]]'
```

使用 `[[ ]]` 是为了避免和 Helm 的 `{{ }}` 模板语法冲突。

### Secret 占位符

必填的 Secret 值使用 `env.Env`：

```yaml
database:
  password: '[[ env.Env.POSTGRES_PASSWORD ]]'
```

`env.Env.POSTGRES_PASSWORD` 会严格读取环境变量。变量不存在时，gomplate 会渲染失败，Pod 不会继续启动。适合数据库密码、用户名、host 这类缺失后配置不可用的值。

可选值或有默认值的配置使用 `getenv`：

```yaml
database:
  port: '[[ getenv "POSTGRES_PORT" "5432" ]]'
  sslmode: '[[ getenv "POSTGRES_SSLMODE" "disable" ]]'
```

`getenv "POSTGRES_PORT" "5432"` 在变量不存在时返回默认值。`getenv "POSTGRES_PASSWORD"` 如果没有默认值且变量不存在，会返回空字符串；因此不建议用它读取必须存在的密码。

`getenv` 还支持 `_FILE` fallback。例如没有设置 `POSTGRES_PASSWORD`，但设置了 `POSTGRES_PASSWORD_FILE=/secrets/postgres-password` 时，`[[ getenv "POSTGRES_PASSWORD" ]]` 会读取该文件内容。`env.Env.POSTGRES_PASSWORD` 不会读取 `_FILE`。

## 环境 Values

每个环境一个 values 文件：

```text
values/
  values-devnet.yaml
  values-testnet.yaml
```

业务配置值统一放在 `appConfig` 下，避免和标准 chart 自身字段冲突：

```yaml
appConfig:
  server:
    port: 8080

  database:
    name: demo_app_devnet
```

## 发布与部署

CI 会把这个目录中的 `files/` 和 `values/` 一起放入最终 config chart 包。

CD 部署时根据目标环境选择 values 文件，例如：

```bash
helm upgrade --install demo-app-config ./demo-app-config-1.2.3.tgz \
  -f values/values-devnet.yaml \
  -n demo
```

如果 CD 从 OCI 或 tgz chart artifact 部署，需要先取得 chart 包中的 `values/values-<env>.yaml`，再作为 Helm `-f` 参数传入。这个 README 只描述业务配置目录约定，不约束具体 CD 实现方式。
