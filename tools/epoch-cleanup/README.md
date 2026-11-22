# 空 epoch 历史清理工具

从 RappaMaster 仓库根目录运行。使用目标 MySQL 对应的应用配置文件，先确认配置中的数据库地址和库名。

## 预览

```bash
go run ./tools/epoch-cleanup -config config.json
```

默认只读。输出中的 `database` 是目标库，`cutoff` 是 30 天前的截止时间，`eligible` 是符合清理条件的行数。`data_bytes`、`index_bytes`、`free_bytes` 是 MySQL 对整张表的空间估算。

候选行必须同时满足：

- `created_at` 早于截止时间。
- `process`、`commits`、`justifieds`、`finalizes` 都是空 JSON 对象，`invalids` 和 `init_tasks` 都是空 JSON 数组。SQL NULL 或 JSON null 不算。
- `t_id` 能匹配 `dev_references` 中同一 epoch 的交易引用。

## 执行清理

先部署支持重建空明细的应用版本，并核对直接读取 `dev_epoches` 的报表。预览结果符合预期后，首次只跑一批：

```bash
go run ./tools/epoch-cleanup -config config.json -apply -max-batches 1
```

每批最多删除 1,000 行，`-max-batches` 默认是 10。输出 `deleted` 是实际删除的行数。清理仅涉及 `dev_epoches` 明细，交易引用和链上记录保留。首次执行后，用旧 epoch ID 和交易哈希抽查详情，再跑一次预览。InnoDB 通常复用删除后腾出的页，数据库文件不一定立即缩小。

### 清除全部符合条件的历史空记录

先运行预览，确认输出中的数据库是目标库。用 `eligible ÷ 1000` 向上取整作为 `-max-batches`。例如，预览显示 `eligible=157928` 时，最多需要 158 批：

```bash
go run ./tools/epoch-cleanup -config /path/to/config.json
go run ./tools/epoch-cleanup -config /path/to/config.json -apply -max-batches 158
go run ./tools/epoch-cleanup -config /path/to/config.json
```

把 `/path/to/config.json` 替换为目标环境的配置路径。如果一轮执行超时，已完成的批次会保留。按重新预览的剩余条数计算批次数并继续，直到 `eligible=0`。如果预览一开始就是 0，可跳过执行。

此命令只覆盖超过 30 天、符合上述判断条件的空明细。最近 30 天内的空明细和有业务内容的记录会保留。
