select
  id            as id,
  instrument_id as instrument_id,
  enabled       as enabled,
  name          as name,
  description   as description,
  protocol      as protocol,
  url           as url
from instruments_controller as c
where
  c.instrument_id = $instrument_id and
  c.name = $name
