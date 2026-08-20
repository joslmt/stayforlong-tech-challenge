# Published OAS inconsistencies

The challenge links to a page containing a current API definition (v3.0.7) and a button that loads a legacy definition (v2.0.1). They describe the same endpoints but disagree in business-significant ways.

| Area | Legacy v2.0.1 / supplied challenge contract | Current v3.0.7 | Project decision |
| --- | --- | --- | --- |
| `/data` metrics | Float, maximum two decimals | Integer | Float, maximum two decimals |
| `/revenue.total_profit` | Float, maximum two decimals | Integer | Float, maximum two decimals |
| `/revenue` night metrics | Float, maximum two decimals | Integer | Float, maximum two decimals |
| `/revenue` metric scope | Selected booking schedule | Descriptions say all booking requests | Selected schedule |
| `/revenue` summary | Outputs selected bookings and their statistics | Maximises night occupation | Maximise occupied nights and report selected statistics |
| `request_ids` ordering | Example appears to preserve request order | Not defined | Chronological check-in order |
| `request_id` format | Free-form string; examples use partner-prefixed IDs | Free-form string | UUID v4, an intentional stricter project rule |

The project follows the decimal schema provided with the challenge because monetary night values naturally produce fractions. It follows the newer definition's explicit occupancy objective, while making tie-breakers and cleaning behavior visible. It deliberately tightens `request_id` to UUID v4; therefore the upstream partner-style ID examples must be replaced before they can be sent to this implementation.

The implementation contract is [`../api/openapi.yaml`](../api/openapi.yaml). It does not pretend the upstream ambiguity does not exist: this document and the ADR explain the compatibility choice reviewers should evaluate.

Sources inspected on 20 August 2026:

- [Current definition v3.0.7](https://interview-fixtures.s3.eu-west-1.amazonaws.com/swe-coding-challenge/index.html)
- [Legacy definition v2.0.1](https://interview-fixtures.s3.eu-west-1.amazonaws.com/swe-coding-challenge/legacy_spec.html)
