resource "bluestonepim_category" "my_category" {
  name = "My category"
}

resource "bluestonepim_matrix_attribute_definition" "size_chart" {
  name    = "Size chart"
  columns = [{ value = "S" }, { value = "M" }]
  rows    = [{ value = "Chest" }]
}

locals {
  columns = bluestonepim_matrix_attribute_definition.size_chart.columns
  rows    = bluestonepim_matrix_attribute_definition.size_chart.rows
}

resource "bluestonepim_category_matrix_attribute" "size_chart" {
  category_id             = bluestonepim_category.my_category.id
  attribute_definition_id = bluestonepim_matrix_attribute_definition.size_chart.id

  # Keyed by column ID, then row ID
  values = {
    (local.columns[0].id) = { (local.rows[0].id) = "92" }
    (local.columns[1].id) = { (local.rows[0].id) = "100" }
  }
}
