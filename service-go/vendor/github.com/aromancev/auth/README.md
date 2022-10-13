## Authentication scheme
```mermaid
sequenceDiagram
    participant User
    participant Web
    participant IAM
    participant Confa
    User->>Web: Open /login
    User->>Web: Submit email
    Web->>IAM: POST /api/iam/v1/login <br/> {email: "<user email>"}
    IAM->>User: Send email with JWT
    User->>Web: Open the link from the email
    Web->>IAM: POST /api/iam/v1/session <br/> Authorisation: Bearer <email JWT> 
    IAM->>IAM: Check email JWT to verify email
    IAM->>Web: Set-Cookie: session=<session key> HttpOnly <br/> {token: <access JWT>, expireIn: <expire seconds>}
    Web->>Confa: POST /api/confa/v1/confas <br/> Authentication: Bearer <access JWT>
    Web->>Web: Wait for refresh
    Web->>IAM: GET /api/iam/v1/token <br/> Cookie: session=<session key>
    IAM->>IAM: Check session by key in DB
    IAM->>Web: {token: <access JWT>, expireIn: <expire seconds>}
```

## Generating keys
Auth uses ECDSA for signing. To generate private / public keys run:
```bash
openssl ecparam -name prime256v1 -genkey -noout -out private-key.pem
openssl ec -in private-key.pem -pubout -out public-key.pem
```
