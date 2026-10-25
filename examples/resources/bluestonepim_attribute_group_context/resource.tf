resource "bluestonepim_context" "dutch" {
  name   = "Dutch"
  locale = "nl-NL"
}

resource "bluestonepim_attribute_group" "general" {
  name = "General"
}

resource "bluestonepim_attribute_group_context" "general_nl" {
  group_id   = bluestonepim_attribute_group.general.id
  context_id = bluestonepim_context.dutch.id
  name       = "Algemeen"
}
