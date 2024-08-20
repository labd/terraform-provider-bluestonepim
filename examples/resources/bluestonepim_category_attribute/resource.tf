resource "bluestonepim_category" "my_category" {
  name = "My category"
}

resource "bluestonepim_attribute_definition" "my_attribute_definition" {
  name      = "My Attribute Definition"
  data_type = "text"
}

resource "bluestonepim_category_attribute" "my_category_attribute" {
  category_id             = bluestonepim_category.my_category.id
  attribute_definition_id = bluestonepim_attribute_definition.my_attribute_definition.id
  mandatory               = true
}
