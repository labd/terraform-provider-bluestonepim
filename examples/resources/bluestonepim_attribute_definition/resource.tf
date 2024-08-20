resource "bluestonepim_attribute_definition" "my_attribute_definition" {
  name         = "My Attribute Definition"
  number       = "my-attribute-definition-key"
  data_type    = "text"
  content_type = "text/markdown"
  description  = "This is a description of the attribute definition."
}
