resource "bluestonepim_context" "dutch" {
  name   = "Dutch"
  locale = "nl-NL"
}

resource "bluestonepim_attribute_definition" "ean" {
  name        = "EAN"
  description = "European Article Number"
  data_type   = "text"
}

resource "bluestonepim_attribute_definition_context" "ean_nl" {
  definition_id = bluestonepim_attribute_definition.ean.id
  context_id    = bluestonepim_context.dutch.id
  name          = "EAN-code"
  description   = "Europees artikelnummer"
}
