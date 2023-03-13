select
  id            as id,
  instrument_id as instrument_id,
  enabled       as enabled,
  type          as type,
  specification as specification
from instruments_automation_job as j
where
  j.enabled = true
