resource "bluestonepim_category" "my_category" {
  name = "My category"
}

resource "bluestonepim_attribute_definition" "amount" {
  name      = "Amount"
  data_type = "decimal"
}

resource "bluestonepim_attribute_definition" "unit" {
  name      = "Unit"
  data_type = "text"
}

resource "bluestonepim_compound_attribute_definition" "quantity" {
  name = "Quantity"
  sub_definitions = [
    { id = bluestonepim_attribute_definition.amount.id },
    { id = bluestonepim_attribute_definition.unit.id },
  ]
}

resource "bluestonepim_category_compound_attribute" "quantity" {
  category_id             = bluestonepim_category.my_category.id
  attribute_definition_id = bluestonepim_compound_attribute_definition.quantity.id
}


# The sub attributes are assigned with the compound attribute, set their values
# with the resource matching their type.
resource "bluestonepim_category_attribute" "unit" {
  category_id             = bluestonepim_category.my_category.id
  attribute_definition_id = bluestonepim_attribute_definition.unit.id
  value                   = "kg"

  depends_on = [bluestonepim_category_compound_attribute.quantity]
}
