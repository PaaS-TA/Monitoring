### cflinuxfs2 BOSH Release

This bosh release contains the rootfs package as well as a job that will
extract the package to `/var/vcap/packages/cflinuxfs2/rootfs`.

#### Trusted certificates

Trusted certs can be installed in the rootfs by setting the
`cflinuxfs2-rootfs.trusted_certs` property to the certificate chain in any
order. For example in your deployment manifest:

```
properties:
  cflinuxfs2-rootfs:
    trusted_certs: |+
      -----BEGIN CERTIFICATE-----
      MIIDQjCCAiqgAwIBAgIJAP/z/IO9Vh6HMA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV
      BAYTAkFVMRMwEQYDVQQIEwpTb21lLVN0YXRlMSEwHwYDVQQKExhJbnRlcm5ldCBX
      aWRnaXRzIFB0eSBMdGQwIBcNMTYwMzAxMTg0NzQ3WhgPMjI4OTEyMTQxODQ3NDda
      MEwxCzAJBgNVBAYTAlVTMRUwEwYDVQQKEwxDTE9VREZPVU5EUlkxJjAkBgNVBAMT
      HWJsb2JzdG9yZS5zZXJ2aWNlLmNmLmludGVybmFsMIIBIjANBgkqhkiG9w0BAQEF
      AAOCAQ8AMIIBCgKCAQEAmcQD0ryug94fllXGM9+mpfHTrT++ZTGZpZ0KCde0iky7
      fprXiVIoHMqgCDnPvSmI7AUZ0TIxYZtm9FfIkdtjk0QW8PbXmbxEBQwH75EgPqNS
      0rkkmwzvVlPI963CTyR0SNpjK8s5GpGO9PQd/OY2AQG1ty1jE1T0YLGdaI2LWsHq
      Y11WFkPYdOfYVnSZeiJkvOkZdZ5KQjZLfgMtyg3cV7yIA552OQiQn3OAFl15K0Bg
      XTjHWfgu5vVGDr/dw0/Dlm2M56EIRBvc/XBJ9/+cj++R3ru77EfJF/h5vn1ahSxZ
      RLWOmDoEOhfDmZ9w/QNWFpHWGJ9dHH7wLN8W/naUkwIDAQABoywwKjAoBgNVHREE
      ITAfgh1ibG9ic3RvcmUuc2VydmljZS5jZi5pbnRlcm5hbDANBgkqhkiG9w0BAQUF
      AAOCAQEABHHiEQW+lG2kM987QeRkycXAwASUtZJV8wrbB8PUn9BC799TVXGkpFLr
      oflUnw4hxAlnwrSYWX+1ueCJ2cR1HoXUMTrwK4a5GUtoHnKJQwnb8K+z7lC+0oqE
      GX+CTb7pPjxwHKGgRw1jjRNLaf3wxnSMrS73OcikF5aGcHylxdCDhAMDvcX6xuqP
      kDMh20Kjg5brpRhZdkBIe9Tja8W4MZfugzZrPamvo14mRinZVnpiXodvhGHxkHvZ
      /SA6HwVuqxoYEC+lqJkLbSqyVPSxLvidKl1Zb6074AZxmVlipmrnEETLaScE98Fp
      XqP9Bw730tiw8W3mZ3NFDQtwAowlqw==
      -----END CERTIFICATE-----
```

#### Requirements

This release depends on recent BOSH versions. In particular it requires bosh
version v206+ (1.3072.0) and stemcell version 3125+.

#### Uploading release to BOSH-Lite

The following command will create and upload the cflinuxfs2 BOSH release
to the BOSH director.

`bosh -n create release && bosh -n upload release`

#### Running smoke tests

1. Add the bundle of trusted certs to `manifest-generation/bosh-lite-stubs/property-overrides.yml`:

```
property_overrides:
  cflinuxfs2-rootfs:
    trusted_certs: |
      -----BEGIN CERTIFICATE-----
      MIIFATCCAuugAwIBAgIBATALBgkqhkiG9w0BAQswEjEQMA4GA1UEAxMHZGllZ29D
      QTAeFw0xNTA3MTYxMzI0MTJaFw0yNTA3MTYxMzI0MTZaMBIxEDAOBgNVBAMTB2Rp
      ZWdvQ0EwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQCsrzEJ5hAQkdkT
      l6z4ffiYvq4RSxKXkeZWTHv5b1w6FSnGCVoQL0ilKyQTGzn001TsZBhqJRmhKvLs
      /4RC8a10KK8hmVhoV4MX690Abd47GbRQR6EPdcd4URHqr0NeeUIPviZGk1EYpFaM
      T81eVq15Q+VrakVfGMjPIPfGqtXV14fs9jvkzVAdTysM8AtZtfwQC3ohVfkL7wA2
      /Xs2YYQdLI1dKNnYdDxaDYmbjjCmxTMlkrloFBLmNveEEpy9Vnw3mcGyuAvq8PEr
      Uua58czKsb81bONp7hzjK8I7BvpvneGTPXg7zzuVRRTwRhZSOoNcqE3/+EjJd5/W
      ONtAYX66xN9apYGHcSmWDFxH6RBwLzJzJOo/FJ0AD5BkQBjJ4x5ZX+5X05oAegj1
      wUYx32q2IrDIJzNF+CltrhY+bhJFmEqy72nomQPowSvuydlJMOYH5ATE8Lww0XzA
      FmhityWvbmrgneSYdg9RvzbqLGTbuEBJ2D+X5WGtAlyvKRehoSJcOr0h9iRCnZIW
      hu9YV6aBsVJHHyc1C4d4cpOx0U5QMXy05Z5wdSQra8n8pG7SC2K9V8HbOidr+4wI
      ZWHwAIgyA0bVvHdGrGeeWeyW/XXD4YGyCAnT4DXWhTLPgxu4gg4rf7nnyHKcAqYp
      DgHKMZOYTnbjCMcXyoYIJ8dR/RvYOQIDAQABo2YwZDAOBgNVHQ8BAf8EBAMCAAYw
      EgYDVR0TAQH/BAgwBgEB/wIBADAdBgNVHQ4EFgQUAU5pu7rUL87RDYhHRL+YYfgc
      /4YwHwYDVR0jBBgwFoAUAU5pu7rUL87RDYhHRL+YYfgc/4YwCwYJKoZIhvcNAQEL
      A4ICAQChLHQM6f769dt9L6MmjOLcYdmmMuyxY8iqdnJIa43MBxKjxzmt6xSIPMBU
      BWFui5gScKPXiA9Nri2Tyzm5zjcQtoJUFcXA8RGgK4aVQ1QCuY4OyiR126WfZiiJ
      J0btSmUXGIme25KEQ2PSiYmwPrLTFG3G+0ylUq6b/rPzHfkFOZXX4U9qLvqY9AnO
      NuYxLT40xDwlL6drcvicEfZ+vV0SABf4HAH+wphRyHR4fkwOBrrieBXvpRUlGeRw
      ZtDVeX8v28WZqoYXV/36JrGbhxSkqBXQk5gdrOUDXebaeQPRvarWCd2zSGmyADei
      npMRDEovA7AlyxX//vBx9MKV3L3NhoL66nBgOwm23DZJLIwCM5AIBvyZMfMpB4sM
      d2nUiXF+5WRFG1bjHuEmU0HvZGXFFzJaiJrnlvzDhJB32DQ5LgEeN+9X42x3DXUZ
      +dR5Qqu0wgQGpdjC9sNsgMBcqVqmc8rWfRxHSusHff7tFs8gpzNRxH6Rimws9M0d
      RFWLAS0T7YSB6deM41Efz7T4Gq+QLm7sv73pDhuIky+AZlWkAr9Wu/+RpNvcQfum
      r5EejEQP82achV3em5+macfNfEIILruStanw9D+kR1GYlE07wMTTmkZ39x3HMicf
      r4ERoMvnaSaiGVHIiCi9ZsoNLlf6TBNNfaqpc8jDZa2/o/nM+Q==
      -----END CERTIFICATE-----
```

2. Run `scripts/generate-bosh-lite-manifest` to generate a BOSH-Lite deployment
   manifest for the smoke test.

3. Run `bosh deployment ./manifests/bosh-lite/rootfs-smoke-test.yml && bosh deploy && bosh run errand cflinuxfs2-smoke-test`
