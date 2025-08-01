resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    gather_facts: false
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
        -----BEGIN OPENSSH PRIVATE KEY-----
        b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAVwAAAAdzc2gtcn
        NhAAAAAwEAAQAAAEEAqPfgaTEWEP3S9w0tgsicURfo+nLW09/0KfOPinhYZ4ouzU+3xC4p
        SlEp8Ut9FgL0AgqNslNaK34Kq+NZjO9DAQAAATB+9/CSfvfwkgAAAAdzc2gtcnNhAAAAQQ
        Co9+BpMRYQ/dL3DS2CyJxRF+j6ctbT3/Qp84+KeFhnii7NT7fELilKUSnxS30WAvQCCo2y
        U1orfgqr41mM70MBAAAAAwEAAQAAAEAgkuLEHLaqkWhLgNKagSajeobLS3rPT0Agm0f7k5
        5FXVt743hwNgkp98bMNrzy9AQ1mJGbQZGrpr4c8ZAx3aRNAAAAIBOs/5OiPgoTdSy7bcF9
        IGpSE8ZgGKzgYQVZeN97YE00AAAAIQCjEr8yAZ54u6Lfzkontk5iS2OEsE0AHr18rBNkWx
        Q2HQAAACEBCUEaRQnMnbp79mxDXDf6AU0cN/RPBjb9qSHDcWZHGzUAAAAXcGhwc2VjbGli
        LWdlbmVyYXRlZC1rZXkBAgME
        -----END OPENSSH PRIVATE KEY-----
        EOT
      }
    ]
  }
}
