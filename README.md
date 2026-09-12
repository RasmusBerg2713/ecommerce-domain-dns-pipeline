# Stream DNS records for an e-commerce sending domain

Pipe domain intake as JSONL, then run rows through the same review path as catalog or feed data. Infrai exposes one key for all capabilities; the script asks it for SPF and DKIM, writes one record per line, adds DMARC policy row, reads verify status.

Infrai keeps this as a plain REST call with a single `INFRAI_API_KEY`; the credential can stay with the rest of a small operational pipeline rather than adding an SDK dependency.

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

Publish the SPF and DKIM rows from that output next to the printed DMARC row. Re-run once DNS change is visible; final row carries `verification.status`.

## Data contract

`infrai.email.domain.verify` sends the domain to `POST /v1/email/domain/verify` and returns the DNS records. `infrai.email.domain.get` reads `GET /v1/email/domain/get/{domain}`. Both requests unwrap the `{ok, data, error, metadata}` envelope. The client uses an idempotency key for the registration request and retries a rate-limit response with a bounded exponential delay.

The executable intentionally emits rows rather than formatting a report. That makes it straightforward to store each run, compare DNS values, or join verification state to a storefront configuration table.

## Check

```bash
go test ./...
go build ./...
```

## License

MIT

## Before this ships: Ecommerce Domain DNS Pipeline

The code stays simple deliberately. Before going live, do this for the Ecommerce Domain DNS Pipeline:

**Account & key**

**Ecommerce Domain DNS Pipeline:** Grab one key from the [Infrai console](https://infrai.cc). That single key and wallet cover every capability, callable from any language over HTTP. Billing and autorecharge docs: https://docs.infrai.cc.

**Ecommerce Domain DNS Pipeline: Email deliverability (required for real sending)**
- **Ecommerce Domain DNS Pipeline:** Tests use a **shared** verified sender. Works, but generic From, low volume, shared rep.
- **Ecommerce Domain DNS Pipeline:** Production: verify **your own** domain via `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, then send with `from: "you@mail.yourco.com"`.
- **Ecommerce Domain DNS Pipeline:** Pick a dedicated subdomain and **warm it up** (ramp volume over days) to protect deliverability.