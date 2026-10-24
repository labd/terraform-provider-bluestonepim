resource "bluestonepim_attribute_group" "dimensions" {
  name          = "Dimensions"
  order_by_name = false
}

resource "bluestonepim_attribute_definition" "width" {
  name      = "Width"
  data_type = "decimal"
  group_id  = bluestonepim_attribute_group.dimensions.id
}

resource "bluestonepim_attribute_definition" "height" {
  name      = "Height"
  data_type = "decimal"
  group_id  = bluestonepim_attribute_group.dimensions.id
}

# Definitions in the group that are not listed here are shown after these
resource "bluestonepim_attribute_group_definition_order" "dimensions" {
  group_id = bluestonepim_attribute_group.dimensions.id
  definition_ids = [
    bluestonepim_attribute_definition.height.id,
    bluestonepim_attribute_definition.width.id,
  ]
}
