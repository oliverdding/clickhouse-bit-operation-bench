# clickhouse-bit-operation-bench

## Structure

```sql
CREATE TABLE IF NOT EXISTS field on cluster default_cluster (
  category LowCardinality(String),
  user_id UUID,
  func_crash UInt64 CODEC(Delta, ZSTD(3)),
  func_lag UInt64 CODEC(Delta, ZSTD(3))
) Engine=MergeTree()
ORDER BY (category)
SETTINGS index_granularity = 8192,
    ttl_only_drop_parts = 1,
    use_minimalistic_part_header_in_zookeeper = 1;

CREATE TABLE IF NOT EXISTS default.field_dist on cluster default_cluster as default.field
ENGINE = Distributed(default_cluster, default, field);

CREATE TABLE IF NOT EXISTS kafka_field on cluster default_cluster (
  category LowCardinality(String),
  user_id UUID,
  func_crash UInt64,
  func_lag UInt64
)engine = Kafka()
settings
kafka_broker_list = '127.0.0.1:9092',
kafka_topic_list = 'test',
kafka_group_name = 'field',
kafka_format = 'JSONEachRow',
kafka_num_consumers = 1,
kafka_client_id = 'clickhouse',
kafka_flush_interval_ms = 1000;

CREATE MATERIALIZED VIEW IF NOT EXISTS mv_field ON cluster default_cluster TO field
AS SELECT
  category,
  user_id,
  func_crash
  func_lag
FROM kafka_field;
```

```sql
CREATE TABLE IF NOT EXISTS bit on cluster default_cluster (
  category LowCardinality(String),
  user_id UUID,
  func_bitmap UInt64 ZSTD(3)
) Engine=MergeTree()
ORDER BY (category)
SETTINGS index_granularity = 8192,
    ttl_only_drop_parts = 1,
    use_minimalistic_part_header_in_zookeeper = 1;

CREATE TABLE IF NOT EXISTS default.bit_dist on cluster default_cluster as default.bit
ENGINE = Distributed(default_cluster, default, bit);

CREATE TABLE IF NOT EXISTS kafka_bit on cluster default_cluster (
  category LowCardinality(String),
  user_id UUID,
  func_crash UInt64,
  func_lag UInt64
)engine = Kafka()
settings
kafka_broker_list = '127.0.0.1:9092',
kafka_topic_list = 'test',
kafka_group_name = 'bit',
kafka_format = 'JSONEachRow',
kafka_num_consumers = 1,
kafka_client_id = 'clickhouse',
kafka_flush_interval_ms = 1000;

CREATE MATERIALIZED VIEW IF NOT EXISTS mv_bit ON cluster default_cluster TO bit
AS SELECT
  category,
  user_id,
  bitAnd(bitShiftLeft(func_lag, 1), func_crash)
FROM kafka_bit;
```
