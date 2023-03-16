update instruments_controller
set
  enabled = $enabled,
  name = $name,
  description = $description,
  protocol = $protocol,
  url = $url
where instruments_controller.id = $id
