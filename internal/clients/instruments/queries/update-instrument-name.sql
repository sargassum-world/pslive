update instruments_instrument
set
  name = $name
where instruments_instrument.id = $id
