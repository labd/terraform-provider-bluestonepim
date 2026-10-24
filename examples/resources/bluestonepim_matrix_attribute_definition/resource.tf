resource "bluestonepim_matrix_attribute_definition" "size_chart" {
  name   = "Size chart"
  number = "size-chart"

  columns = [
    { value = "S" },
    { value = "M" },
    { value = "L" },
  ]

  rows = [
    { value = "Chest" },
    { value = "Waist" },
  ]
}
