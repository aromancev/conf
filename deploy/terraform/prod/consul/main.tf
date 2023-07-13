resource "tls_private_key" "main" {
  algorithm   = "ECDSA"
  ecdsa_curve = "P256"
}

resource "consul_key_prefix" "auth" {
  path_prefix = "auth/"

  subkeys = {
    "private_key" = tls_private_key.main.private_key_pem
    "public_key"  = tls_private_key.main.public_key_pem
  }
}

resource "consul_key_prefix" "web" {
  path_prefix = "web/"

  subkeys = {
    "host"   = "confa.io"
    "scheme" = "https"
  }
}

resource "consul_key_prefix" "beanstalk_tubes" {
  path_prefix = "beanstalk/"

  subkeys = {
    "tubes/send"                   = "sender/send"
    "tubes/update-avatar"          = "confa/update-avatar"
    "tubes/start-recording"        = "confa/start-recording"
    "tubes/stop-recording"         = "confa/stop-recording"
    "tubes/recording-update"       = "confa/recording-update"
    "tubes/store-event"            = "rtc/store-event"
    "tubes/update-recording-track" = "rtc/update-recording-track"
    "tubes/process-track"          = "avp/process-track"
  }
}

resource "consul_key_prefix" "storage" {
  path_prefix = "storage/"

  subkeys = {
    "public-url"                    = "/api/storage"
    "buckets/user-public"           = "user-public"
    "buckets/user-uploads"          = "user-uploads"
    "buckets/confa-tracks-internal" = "confa-tracks-internal"
    "buckets/confa-tracks-public"   = "confa-tracks-public"
  }
}
