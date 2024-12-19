provider "echo" {
  data = ephemeral.ansible_navigator_run.test
}

resource "echo" "test" {}
