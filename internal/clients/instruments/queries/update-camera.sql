update instruments_camera
set
  enabled = $enabled,
  protocol = $protocol,
  url = $url
where instruments_camera.id = $id
