resource "bluestonepim_relation" "successor" {
  name         = "Superseded by"
  reverse_name = "Successor of"
  number       = "successor"
  description  = "The product that replaces this one."
  direction    = "TWO_WAY"
}
