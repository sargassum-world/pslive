update instruments_camera
set
  url = $url,
  protocol = $protocol
where instruments_camera.id = $id
