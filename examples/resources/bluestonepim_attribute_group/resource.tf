resource "bluestonepim_attribute_group" "dimensions" {
  name   = "Dimensions"
  number = "dimensions"

  # Order the definitions manually, see bluestonepim_attribute_group_definition_order
  order_by_name = false
}

resource "bluestonepim_attribute_definition" "width" {
  name      = "Width"
  data_type = "decimal"
  group_id  = bluestonepim_attribute_group.dimensions.id
}
