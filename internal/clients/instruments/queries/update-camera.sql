update instruments_camera
set
  url = $url,
  protocol = $protocol,
  enabled = $enabled
where instruments_camera.id = $id
