update instruments_instrument
set
  description = $description
where instruments_instrument.id = $id
