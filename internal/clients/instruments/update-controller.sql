update instruments_controller
set
  url = $url,
  protocol = $protocol
where instruments_controller.id = $id
