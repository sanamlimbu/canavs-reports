locals {
  service_name = "canvas-report"

  # for runtime "provided.al2023" binary name must be "bootstrap"
  binary_name  = "bootstrap"
  binary_path  = "${path.module}/tf_generated/${local.binary_name}"
  archive_path = "${path.module}/tf_generated/${local.service_name}.zip"
}