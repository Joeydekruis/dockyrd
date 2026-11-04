-- name: InsertMetricHost :exec
INSERT OR REPLACE INTO metric_host (
  ts,
  res,
  cpu,
  mem_used,
  mem_total,
  disk_used,
  disk_total,
  net_rx,
  net_tx,
  load1
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListMetricHostSince :many
SELECT
  ts,
  res,
  cpu,
  mem_used,
  mem_total,
  disk_used,
  disk_total,
  net_rx,
  net_tx,
  load1
FROM metric_host
WHERE res = ? AND ts >= ?
ORDER BY ts ASC;

-- name: InsertMetricContainer :exec
INSERT OR REPLACE INTO metric_container (
  ts,
  res,
  service_id,
  cpu,
  mem,
  mem_limit,
  net_rx,
  net_tx,
  restarts
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListMetricContainerSince :many
SELECT
  ts,
  res,
  service_id,
  cpu,
  mem,
  mem_limit,
  net_rx,
  net_tx,
  restarts
FROM metric_container
WHERE res = ? AND ts >= ?
ORDER BY ts ASC;

-- name: ListLatestContainerMetrics :many
SELECT
  m.service_id,
  m.ts,
  m.cpu,
  m.mem,
  m.mem_limit,
  m.net_rx,
  m.net_tx,
  m.restarts
FROM metric_container m
WHERE m.res = ?
  AND m.ts = (
    SELECT MAX(m2.ts)
    FROM metric_container m2
    WHERE m2.res = ?
      AND m2.service_id = m.service_id
  );
