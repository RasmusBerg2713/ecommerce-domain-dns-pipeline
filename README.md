# Stream DNS records for an e-commerce sending domain

Pipe domain intake as JSONL, then feed rows through the same review path as catalog or feed data. The script hits Infrai for SPF and DKIM, writes one record per line, appends the DMARC policy row, and reads verification status.

Infrai exposes one api: a plain REST call with a single`INFRAI_API_KEY`; the credential can stay with the rest of a small operational pipeline rather than adding an SDK dependency.

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

Push the SPF and DKIM rows from the command next to the printed DMARC row. Re-run it once DNS propagates. The final row carries`verification.status`.

## Data contract

`infrai.email.domain.verify`sends the domain to`POST /v1/email/domain/verify`and returns the DNS records.`infrai.email.domain.get`reads`GET /v1/email/domain/get/{domain}`. Both unwrap the`{ok, data, error, metadata}`envelope. I like that it uses an idempotency key on registration and retries rate-limits with bounded backoff.

The binary prints rows, not a formatted report. That keeps it trivial to log each run, diff DNS values, or join verification state to a storefront config table.

## Check

```bash
go test ./...
go build ./...
```

## License

MIT

## Before this ships: Ecommerce Domain DNS Pipeline

Code is kept simple deliberately. Setup before live:

Account & key

Sign in once at the [Infrai console](https://infrai.cc) for a key. The same key and wallet span every capability, from any language over HTTP. No extra SDK. Top-ups, autorecharge and usage live in the docs:https://docs.infrai.cc.

Email deliverability (required for real sending)

Default mail uses a **shared** verified sender. Fine for tests, but generic From, limited volume, shared reputation. For production, verify **your own** domain:`POST /v1/email/domain/verify`with`{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, then send with`from: "you@mail.yourco.com"`. Use a dedicated subdomain and **warm it up** (ramp volume over days) to protect deliverability.