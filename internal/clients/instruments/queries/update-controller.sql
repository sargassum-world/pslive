update instruments_controller
set
  enabled = $enabled,
  protocol = $protocol,
  url = $url
where instruments_controller.id = $id
