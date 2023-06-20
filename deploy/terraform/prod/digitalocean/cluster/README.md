## Generating Secrets

Generate gossip key for Consul or Nomad:
```bash
openssl rand -base64 32
```

Generate certificates for Consul and Nomad:
```bash
consul tls ca create
consul tls cert create -server -dc dc1 -domain consul

nomad tls ca create
nomad tls cert create -server -region global
nomad tls cert create -client
```
