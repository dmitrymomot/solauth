# solauth

[![Tests](https://github.com/dmitrymomot/solauth/actions/workflows/tests.yml/badge.svg)](https://github.com/dmitrymomot/solauth/actions/workflows/tests.yml)

Issuing access token based on signed message by Solana wallet.

## Usage

### 1. Start server

```bash
$ ./bin/server
```

### 2. Request authorization

```bash
$ curl -X POST -H "Content-Type: application/json" -d '{"public_key": "[base58 encoded wallet address]"}' http://localhost:8080/auth/request
```

### 3. Sign message

Sign the message with your wallet.

### 4. Get access token

```bash
$ curl -X POST -H "Content-Type: application/json" -d '{"public_key": "[base58 encoded wallet address]", "signature": "[base64 encoded signature]", "message": "[same message from first request]"}' http://localhost:8080/auth/verify
```

### 5. Refresh access token

```bash
$ curl -X POST -H "Content-Type: application/json" -d '{"refresh_token": "[refresh token from prev req]"}' http://localhost:8080/auth/refresh
```
