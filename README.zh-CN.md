<p align="right">
  <a href="./README.md">English</a>
</p>

<div align="center">

# PoolX

基于 Gin 和 React 的代理池控制面，核心目标是把 Clash/Mihomo 节点组织成可复用的代理池，服务爬虫、抓取、代理请求和自动化出网场景。

</div>

> [!NOTE]
> 使用 glider 进行的节点转代理池已过时，旧版存档在 `glider` 分支。当前项目提供一整个图形界面系统用于构建代理池，暂仅支持 Mihomo 内核。

> [!WARNING]
> 本项目仅为代理池控制面，不提供节点以及其获取方式。本项目仅供学习、研究与技术交流使用，禁止用于任何非法用途。

> [!WARNING]
> 使用初始管理员账号首次登录后，务必立即修改默认密码！

## 项目介绍

在使用爬虫对目标网站进行多次请求时，往往会触发反爬机制，导致访问被封禁。此类封禁通常基于 IP 地址，因此通过切换 IP 可以有效规避限制，从而实现持续稳定的爬取。

常见的解决方案是使用代理池。但现有方案中：
* 免费代理池质量普遍较低，稳定性和可用性难以保障
* 付费代理服务成本较高，往往超出实际需求

一种更具性价比的方案，是利用"机场"提供的节点资源，结合开源代理内核，将其构建为可复用的代理池。

本项目提供了一个简洁易用的 UI 界面，用于高效管理和组织代理节点，无需手动编写或维护复杂的配置文件。同时，项目还提供统一的内核控制能力，便于对代理核心进行集中管理。

### 核心功能

* 支持通过配置文件或订阅地址导入和管理大量节点
* 将节点组织为可复用的代理池
* 为爬虫、自动化任务及代理请求系统提供稳定的本地代理入口
* 支持负载均衡、自动回退以及基于延迟的节点选择策略
* 提供可视化控制面板，用于管理内核运行状态
* 配置可用时应用启动自动拉起内核

## 功能特性

* Web 管理端，内置登录鉴权、系统设置能力
* 节点池管理，支持节点测试、筛选和跨工作台复用
* 基于工作台的端口配置管理，用于构建代理池入口
* 监听多个端口，每一个端口对应不同的代理池配置
* 内置 zashboard，进一步拓展内核监控和管理能力

## 典型使用场景

* 爬虫和抓取服务的代理池
* 自动化平台的统一出网代理网关
* 数据采集任务的轮换代理或回退代理入口
* 面向不同目标站点的工作台式代理编排
* 代理访问 AI，降低网站风险，提升访问稳定性

## 快速开始

### Docker 部署

```bash
cp config.example.yaml config.yaml
# 填写 secrets.jwtSecret、secrets.credentialEncryptionKey、bootstrapAdmin.password
docker compose up -d
docker compose logs -f foam
```

后端默认监听：`http://127.0.0.1:8000`

### 本地开发

```bash
cp config.example.yaml config.yaml
# 修改密钥后：
make dev              # 前端 :8010 + 后端 :8000
```

或分终端启动：

```bash
cd backend && go run ./cmd/foam --config ../config.yaml
cd frontend && pnpm install && pnpm dev
```

开发 UI：`http://127.0.0.1:8010` → 反代 API → `http://127.0.0.1:8000`

## 配置说明

PoolX 不要求你手工维护一个固定的配置文件。

它的核心流程是：

* 导入节点
* 组织节点池
* 定义工作台监听配置
* 自动渲染最终内核运行配置

复制模板并填写必要密钥：

```bash
cp config.example.yaml config.yaml
```

必填字段：
- `secrets.jwtSecret` — 至少 32 个字符
- `secrets.credentialEncryptionKey` — base64 编码的 32 字节密钥
- `bootstrapAdmin.password` — 初始管理员密码

运行参数、部署方式和系统配置请参考：

* [docs/architecture.md](./docs/architecture.md)
* [docs/backend-development.md](./docs/backend-development.md)

## License

MIT — 见 [LICENSE](./LICENSE)。
