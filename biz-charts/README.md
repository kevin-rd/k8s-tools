# biz-charts

通用业务 Helm Charts。

当前包含：

- `go`: Go 服务模板
- `static`: 静态站点模板

## 使用方式

### 1. 本地渲染

```bash
helm template demo ./biz-charts/static
helm template demo ./biz-charts/go
```

带自定义 values：

```bash
helm template demo ./biz-charts/static -f ./biz-charts/static/values.example.yaml
```

### 2. 本地安装

```bash
helm install demo ./biz-charts/static -n demo --create-namespace
```

### 3. 从 GHCR 安装

```bash
helm install demo oci://ghcr.io/<owner>/biz-charts/static --version <chart-version> -n demo --create-namespace
```

例如：

```bash
helm install demo oci://ghcr.io/<owner>/biz-charts/go --version 0.1.0 -n demo --create-namespace
```

### 4. 拉取并查看

```bash
helm pull oci://ghcr.io/<owner>/biz-charts/static --version <chart-version>
helm show values oci://ghcr.io/<owner>/biz-charts/static --version <chart-version>
```

## 发布说明

`biz-charts` 下的 chart 会通过 GitHub Actions 发布到 GHCR。

- 触发方式：push `v*` tag
- 发布范围：仅发布当前 tag 对应提交中发生变更的 chart
- 目标仓库：`ghcr.io/<owner>/biz-charts`

修改 chart 后，记得同步更新对应 `Chart.yaml` 中的 `version`。
