update instruments_automation_job
set
  enabled = $enabled,
  name = $name,
  description = $description,
  type = $type,
  specification = $specification
where instruments_automation_job.id = $id
