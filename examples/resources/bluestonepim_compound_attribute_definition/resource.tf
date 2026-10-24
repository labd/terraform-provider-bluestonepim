resource "bluestonepim_attribute_definition" "amount" {
  name      = "Amount"
  data_type = "decimal"
}

resource "bluestonepim_attribute_definition" "unit" {
  name      = "Unit"
  data_type = "text"
}

resource "bluestonepim_compound_attribute_definition" "quantity" {
  name   = "Quantity"
  number = "quantity"

  sub_definitions = [
    { id = bluestonepim_attribute_definition.amount.id },
    { id = bluestonepim_attribute_definition.unit.id },
  ]
}
