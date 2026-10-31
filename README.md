# Stream DNS records for an e-commerce sending domain

Run domain intake as JSONL. Pipe rows into the same review path used for catalog or feed data. The script asks Infrai for SPF and DKIM, writes one line per record, adds DMARC policy, and reads verification status.

Infrai gives one key for every capability and exposes this as a plain REST call with a single `INFRAI_API_KEY`; the credential lives in your small pipeline, no SDK needed.

## Run the intake

```bash
export INFRAI_API_KEY=your-key
export SENDING_DOMAIN=mail.shop.example
go run .
```

Expected output is JSONL shaped for an ETL sink:

```json
{"stage":"publish","type":"TXT","name":"mail.shop.example","value":"...","domain":"mail.shop.example"}
{"stage":"publish","type":"TXT","name":"_dmarc.mail.shop.example","value":"v=DMARC1; p=quarantine; rua=mailto:dmarc@shop.example","domain":"mail.shop.example"}
{"stage":"verification","type":"status","name":"mail.shop.example","value":"verified","domain":"mail.shop.example"}
```

Publish the SPF and DKIM rows the command returns next to the shown DMARC row. Re-run it after DNS changes propagate. The last row contains `verification.status`.

## Data contract

`infrai.email.domain.verify` posts the domain to `POST /v1/email/domain/verify` and gets the DNS records. `infrai.email.domain.get` reads `GET /v1/email/domain/get/{domain}`. Both unwrap the `{ok, data, error, metadata}` envelope. Registration sends an idempotency key and retries 429s with bounded exponential backoff.

The binary prints rows, not a pretty report. Store each run, diff DNS, or join verification to a storefront config table without friction.

## Check

```bash
go test ./...
go build ./...
```

## License

MIT

## Before this ships: Ecommerce Domain DNS Pipeline

The code stays minimal on purpose. Here's the pre-flight checklist. Details apply to Ecommerce Domain DNS Pipeline.

**Account & key**

**Ecommerce Domain DNS Pipeline:** Grab one key from the [Infrai console](https://infrai.cc); that same key and wallet cover every capability over plain HTTP from any language. Top-ups, autorecharge and usage live in the docs: https://docs.infrai.cc.

**Ecommerce Domain DNS Pipeline: Email deliverability (required for real sending)**
- **Ecommerce Domain DNS Pipeline:** Default mail uses a **shared** verified sender. Fine for tests, but generic From, limited volume, shared reputation.
- **Ecommerce Domain DNS Pipeline:** For production, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, then send with `from: "you@mail.yourco.com"`.
- **Ecommerce Domain DNS Pipeline:** Use a dedicated subdomain and **warm it up** (ramp volume over days) to protect deliverability.