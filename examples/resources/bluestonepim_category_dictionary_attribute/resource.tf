resource "bluestonepim_category" "my_category" {
  name = "My category"
}

resource "bluestonepim_dictionary_attribute_definition" "country" {
  name = "Country"
}

resource "bluestonepim_dictionary_attribute_value" "netherlands" {
  attribute_definition_id = bluestonepim_dictionary_attribute_definition.country.id
  value                   = "Netherlands"
}

resource "bluestonepim_category_dictionary_attribute" "country" {
  category_id             = bluestonepim_category.my_category.id
  attribute_definition_id = bluestonepim_dictionary_attribute_definition.country.id
  value_ids               = [bluestonepim_dictionary_attribute_value.netherlands.id]
  mandatory               = true
}
