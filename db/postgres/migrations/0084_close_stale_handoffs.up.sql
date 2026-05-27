-- 0084_close_stale_handoffs
-- One-time backfill: close handoffs that were stranded in active states
-- (pending / dispatched / running) by the silent-handoff bug. Older
-- TriggerHandoff implementations only pushed a handoff to a terminal
-- state when the target bot called issue_comment. If the bot chose to
-- stay silent the row stayed active forever and the issue UI rendered
-- "X is running" indefinitely.
--
-- The companion code fix in TriggerHandoff prevents new instances of
-- this state; this migration cleans up rows that pre-date the fix.
-- We deliberately limit cleanup to rows older than 10 minutes so any
-- handoff still legitimately in flight at deploy time is not touched.

UPDATE agent_handoffs
SET
  status = 'completed',
  failure_reason = CASE
    WHEN failure_reason = '' THEN 'silent-close-backfill'
    ELSE failure_reason
  END,
  completed_at = COALESCE(completed_at, now()),
  updated_at = now()
WHERE status IN ('pending', 'dispatched', 'running')
  AND updated_at < now() - INTERVAL '10 minutes';
