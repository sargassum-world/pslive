update instruments_camera
set
  enabled = $enabled,
  name = $name,
  description = $description,
  protocol = $protocol,
  url = $url
where instruments_camera.id = $id
