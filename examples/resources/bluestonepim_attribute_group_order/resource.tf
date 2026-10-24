resource "bluestonepim_attribute_group" "general" {
  name = "General"
}

resource "bluestonepim_attribute_group" "dimensions" {
  name = "Dimensions"
}

# Groups not listed here are shown after these
resource "bluestonepim_attribute_group_order" "this" {
  group_ids = [
    bluestonepim_attribute_group.general.id,
    bluestonepim_attribute_group.dimensions.id,
  ]
}
