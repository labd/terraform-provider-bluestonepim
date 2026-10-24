resource "bluestonepim_column_attribute_definition" "dimensions" {
  name        = "Dimensions"
  number      = "dimensions"
  description = "Dimensions of the product"

  columns = [
    { value = "Width" },
    { value = "Height" },
    { value = "Depth" },
  ]
}
