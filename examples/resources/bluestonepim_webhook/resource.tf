resource "bluestonepim_webhook" "my_webhook" {
  url    = "https://example.test"
  secret = "my-secret"
  active = true
  event_types = [
    "PRODUCT_CREATED",
    "PRODUCT_SYNC_DONE"
  ]
}
