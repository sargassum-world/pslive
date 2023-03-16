select
  i.id                as id,
  i.name              as name,
  i.description       as description,
  i.admin_identity_id as admin_id,
  ca.id               as camera_id,
  ca.enabled          as camera_enabled,
  ca.name             as camera_name,
  ca.description      as camera_description,
  ca.protocol         as camera_protocol,
  ca.url              as camera_url,
  co.id               as controller_id,
  co.enabled          as controller_enabled,
  co.name             as controller_name,
  co.description      as controller_description,
  co.protocol         as controller_protocol,
  co.url              as controller_url,
  aj.id               as automation_job_id,
  aj.enabled          as automation_job_enabled,
  aj.name             as automation_job_name,
  aj.description      as automation_job_description,
  aj.type             as automation_job_type,
  aj.specification    as automation_job_specification
from instruments_instrument as i
left join instruments_camera as ca
  on i.id = ca.instrument_id
left join instruments_controller as co
  on i.id = co.instrument_id
left join instruments_automation_job as aj
  on i.id = aj.instrument_id
where
  i.admin_identity_id = $admin_id
