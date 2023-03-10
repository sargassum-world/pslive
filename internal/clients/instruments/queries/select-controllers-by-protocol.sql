select
  id            as id,
  instrument_id as instrument_id,
  url           as url,
  protocol      as protocol,
  enabled       as enabled
from instruments_controller as c
where
  c.protocol = $protocol
