resource "bluestonepim_dictionary_attribute_definition" "country" {
  name                  = "Country"
  number                = "country"
  selected_values_limit = 1
}

resource "bluestonepim_dictionary_attribute_definition" "city" {
  name                 = "City"
  number               = "city"
  parent_definition_id = bluestonepim_dictionary_attribute_definition.country.id
}
