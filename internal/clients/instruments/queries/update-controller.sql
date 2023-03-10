update instruments_controller
set
  url = $url,
  protocol = $protocol,
  enabled = $enabled
where instruments_controller.id = $id
