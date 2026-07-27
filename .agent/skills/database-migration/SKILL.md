---
name: "database-migration"
description: "Foam 项目专用：新增或修改数据库表结构、索引、goose SQL 迁移、internal/infra/persistence/migrator 时必须使用。指导在 migrator/goose 下编写 PostgreSQL/SQLite 双方言 SQL，禁止生产路径 GORM AutoMigrate。"
---

# Foam 数据库迁移（goose + SQL）

Foam 使用 [`github.com/pressly/goose/v3`](https://github.com/pressly/goose) 执行 **嵌入式 SQL 迁移**，方式对齐 Wavelet：

- 入口：`migrator.Up(ctx, gormDB, driver)`  
  由 `relational.Database.InitializeSchema` → `app.New` 在启动时调用
- SQL 嵌入：`//go:embed goose/postgres/*.sql goose/sqlite/*.sql`
- 版本表：默认 `goose_db_version`
- **无** ClickHouse；仅 **SQLite + PostgreSQL**

## 目录

```text
backend/internal/infra/persistence/migrator/
  migrator.go
  migrator_test.go
  goose/
    postgres/
      202607240001_initial_schema.sql
      ...
    sqlite/
      202607240001_initial_schema.sql   # 同版本号、同文件名
      ...
```

Go 模型仍在 `relational/models.go` 等处，**仅做读写映射**；表结构变更必须写 goose SQL。

## 基本规则

1. **PostgreSQL 与 SQLite 必须同一版本号、同一语义文件名**。
2. 迁移标记：

```sql
-- +goose Up
...

-- +goose Down
...
```

3. **不要**在生产路径使用 GORM `AutoMigrate` 改表。
4. **不要**把新表结构只写在 Go model 里而不写 SQL。
5. **DDL 与 seed/DML 分文件**：先结构，再（若需要）下一版本插数据。
6. **禁止物理外键**；关系用显式索引（与 Wavelet 一致）。
7. 管理员默认账号仍由 `adminauth.Bootstrap` + `config.yaml` 创建（**不要**把密码 seed 进 SQL）。
8. SQLite 复杂改表用：

```sql
-- +goose StatementBegin
...
-- +goose StatementEnd
```

## 方言对照

| 关注点 | PostgreSQL | SQLite |
| :--- | :--- | :--- |
| 自增主键 | `BIGSERIAL` | `INTEGER PRIMARY KEY AUTOINCREMENT` |
| 时间 | `TIMESTAMPTZ` | `DATETIME` |
| JSON | `JSONB` 或 `TEXT` | `JSON` / `TEXT` |
| 幂等建表 | `CREATE TABLE IF NOT EXISTS` | 同左 |

## 新增迁移流程

1. 确认涉及的 GORM model、repository、API。
2. 选择下一版本号：`YYYYMMDDNNNN`（例：`202607250001`）。
3. 在 `goose/postgres/` 与 `goose/sqlite/` **各建同名文件**。
4. 编写 `Up` / `Down`（可安全回滚时写 Down）。
5. 同步更新 `relational/models.go`（及 mapping）列映射。
6. 验证：

```bash
cd backend
go test ./internal/infra/persistence/migrator/...
go test ./...
```

7. 若改了 API：按需 `make swagger`；收工前 `make format` / `make code-check`。

## 禁止事项

| 错误 | 正确 |
| :--- | :--- |
| 只改 model 指望 AutoMigrate | goose SQL 双方言 + model 对齐 |
| 只写 postgres 不写 sqlite | 两个目录同版本 |
| 在 handler/repository 里 `CREATE TABLE` | 只在 goose SQL |
| 在 SQL 里写死 bootstrap 密码 | 用 `bootstrapAdmin` 配置 |

## 验证重点

- 空库上 `Up` 后存在：`admins`、`admin_sessions`、`examples`、`runtime_settings`、`goose_db_version`
- 二次 `Up` 幂等（`Applied == false`）
- model 列名/类型与 SQL 一致
- 现有测试通过 `InitializeSchema` → goose（无需 AutoMigrate）

## 与 new-api 技能的关系

新增业务资源时：

1. 写 goose 迁移建表/索引（本技能）
2. 复制 `example` 垂直切片并接线（`new-api`）
