resource "bluestonepim_category" "my_parent_category" {
  name   = "My parent category"
  number = "my-parent-category-key"
}

resource "bluestonepim_category" "my_category" {
  name      = "My category"
  number    = "my-category-key"
  parent_id = bluestonepim_category.my_parent_category.id
}
