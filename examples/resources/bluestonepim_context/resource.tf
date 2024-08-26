resource "bluestonepim_context" "nl_nl" {
  name        = "Dutch (Netherlands)"
  locale      = "nl-NL"
  fallback_id = "en"
}

resource "bluestonepim_context" "nl_be" {
  name        = "Dutch (Belgium)"
  locale      = "nl-BE"
  fallback_id = bluestonepim_context.nl_nl.id
}
