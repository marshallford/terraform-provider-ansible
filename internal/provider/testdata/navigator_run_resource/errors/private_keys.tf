resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    become: false
  EOT
  inventory                = "# localhost"
  ansible_options = {
    private_keys = [
      {
        name = "public"
        data = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAQQCo9+BpMRYQ/dL3DS2CyJxRF+j6ctbT3/Qp84+KeFhnii7NT7fELilKUSnxS30WAvQCCo2yU1orfgqr41mM70MB phpseclib-generated-key"
      },
      {
        name = "encrypted"
        data = <<-EOT
        -----BEGIN OPENSSH PRIVATE KEY-----
        b3BlbnNzaC1rZXktdjEAAAAACmFlczI1Ni1jdHIAAAAGYmNyeXB0AAAAGAAAABCmZ5U5Eu
        qcHFCIfF9gfNrvAAAAEAAAAAEAAABXAAAAB3NzaC1yc2EAAAADAQABAAAAQQCo9+BpMRYQ
        /dL3DS2CyJxRF+j6ctbT3/Qp84+KeFhnii7NT7fELilKUSnxS30WAvQCCo2yU1orfgqr41
        mM70MBAAABMM5HiDWh0Vf5BWUVoso9wTFYoNtxPBfHa3NQk+i/1XLx0ZQbYfurzUkE54Zi
        gVPaGYMHbK1whuxSmRD6JlI3w+lENdjgTX/PR6netDsYKO0AbFxKEmzAItGbJAekcqdELA
        QjEvPDO6BQaBKrBNqrj9S21uA7pNZyqV6uZL7pSZR89B8OmLpN5v5IzXFkjmYzwpT71b+C
        gZ0q2mOH60b+1h1mN3jFjLPVIrpUiUzDhscX6hjd1XD3a69CjsN5IKUbEVp0zb4QNCz7RU
        a4fY8yTCwSQ3VBloX1+ysKv/OX75Bb4WtLpUz3V/gehiYuY9Qm4Cq9wfXI3WgBqFld/8z+
        qmrujXsdNGHAGaHCLD5TQLOn3ZBpEzfLBLcOka89zUAs+JDfHOB/UJ3T1raVNriM3Gc=
        -----END OPENSSH PRIVATE KEY-----
        EOT
      },
      {
        name = "!invalid"
        data = <<-EOT
        -----BEGIN RSA PRIVATE KEY-----
        MIIBOgIBAAJBAKj34GkxFhD90vcNLYLInFEX6Ppy1tPf9Cnzj4p4WGeKLs1Pt8Qu
        KUpRKfFLfRYC9AIKjbJTWit+CqvjWYzvQwECAwEAAQJAIJLixBy2qpFoS4DSmoEm
        o3qGy0t6z09AIJtH+5OeRV1be+N4cDYJKffGzDa88vQENZiRm0GRq6a+HPGQMd2k
        TQIhAKMSvzIBnni7ot/OSie2TmJLY4SwTQAevXysE2RbFDYdAiEBCUEaRQnMnbp7
        9mxDXDf6AU0cN/RPBjb9qSHDcWZHGzUCIG2Es59z8ugGrDY+pxLQnwfotadxd+Uy
        v/Ow5T0q5gIJAiEAyS4RaI9YG8EWx/2w0T67ZUVAw8eOMB6BIUg0Xcu+3okCIBOs
        /5OiPgoTdSy7bcF9IGpSE8ZgGKzgYQVZeN97YE00
        -----END RSA PRIVATE KEY-----
        EOT
      }
    ]
  }
}
