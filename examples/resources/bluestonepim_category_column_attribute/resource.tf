resource "bluestonepim_category" "my_category" {
  name = "My category"
}

resource "bluestonepim_column_attribute_definition" "dimensions" {
  name = "Dimensions"
  columns = [
    { value = "Width" },
    { value = "Height" },
  ]
}

resource "bluestonepim_category_column_attribute" "dimensions" {
  category_id             = bluestonepim_category.my_category.id
  attribute_definition_id = bluestonepim_column_attribute_definition.dimensions.id
  propagate               = true

  values = {
    (bluestonepim_column_attribute_definition.dimensions.columns[0].id) = "10"
    (bluestonepim_column_attribute_definition.dimensions.columns[1].id) = "20"
  }
}
