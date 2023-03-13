update instruments_automation_job
set
  enabled = $enabled,
  type = $type,
  specification = $specification
where instruments_automation_job.id = $id
