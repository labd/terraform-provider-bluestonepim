resource "bluestonepim_attribute_group" "dimensions" {
  name   = "Dimensions"
  number = "dimensions"
}

resource "bluestonepim_attribute_definition" "width" {
  name      = "Width"
  data_type = "decimal"
  group_id  = bluestonepim_attribute_group.dimensions.id
}
