select
  id            as id,
  instrument_id as instrument_id,
  enabled       as enabled,
  protocol      as protocol,
  url           as url
from instruments_controller as c
where
  c.protocol = $protocol and
  c.enabled = true
