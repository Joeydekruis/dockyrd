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
