select
  i.id            as id,
  i.name          as name,
  i.description   as description,
  i.admin_user_id as admin_id,
  ca.id           as camera_id,
  ca.url          as camera_url,
  ca.protocol     as camera_protocol,
  co.id           as controller_id,
  co.url          as controller_url,
  co.protocol     as controller_protocol
from instruments_instrument as i
left join instruments_camera as ca
  on i.id = ca.instrument_id
left join instruments_controller as co
  on i.id = co.instrument_id
-- TODO: add pagination
