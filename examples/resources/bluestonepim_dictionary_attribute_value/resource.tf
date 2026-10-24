resource "bluestonepim_dictionary_attribute_definition" "country" {
  name = "Country"
}

resource "bluestonepim_dictionary_attribute_value" "netherlands" {
  attribute_definition_id = bluestonepim_dictionary_attribute_definition.country.id
  value                   = "Netherlands"
  number                  = "nl"
}
