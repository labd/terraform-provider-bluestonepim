resource "bluestonepim_context" "dutch" {
  name   = "Dutch"
  locale = "nl-NL"
}

resource "bluestonepim_category" "bikes" {
  name        = "Bikes"
  description = "All our bikes"
}

resource "bluestonepim_category_context" "bikes_nl" {
  category_id = bluestonepim_category.bikes.id
  context_id  = bluestonepim_context.dutch.id
  name        = "Fietsen"
  description = "Al onze fietsen"
}
