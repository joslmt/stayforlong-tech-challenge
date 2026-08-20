# Business decisions

The challenge intentionally leaves room for judgement. This document makes that judgement visible instead of hiding it in code.

## Booking inputs

- The request contains between 1 and 10,000 bookings and is limited to 5 MiB.
- `request_id` is a hyphenated UUID v4. Hexadecimal input is case-insensitive, normalised to lowercase, and unique within the request.
- Two bookings may have identical commercial details when they have different IDs.
- `check_in` is a real calendar date in `YYYY-MM-DD` format. Historical and future dates are accepted because both endpoints calculate, rather than create, bookings.
- `nights` is between 1 and 365.
- `selling_rate` is a positive whole-euro amount for the complete stay.
- `margin` is an integer percentage between 0 and 100 inclusive. A zero-margin booking may be selected when it improves occupancy.
- Unknown JSON fields are rejected to catch client mistakes early.
- The request concerns one apartment and no pre-existing confirmed bookings. Taxes, fees, currency conversion, authentication, and persistence are outside the challenge.

## Profit and statistics

Whole-stay profit is:

```text
selling rate × margin percentage
```

The average night value is total profit divided by total nights, so longer stays are weighted by their occupied nights. Minimum and maximum compare each booking's profit per night.

Values are represented exactly as cents and ratios. Returned EUR numbers are rounded half-up to two decimal places only when serialising the response. This avoids cumulative errors from rounded intermediate values.

`/data` statistics cover every input booking. `/revenue` statistics and `total_profit` cover only selected bookings.

## Scheduling

The primary objective is maximum occupied nights—not booking count or profit. Equal-occupancy schedules prefer:

1. higher exact total profit;
2. the schedule containing the earliest original input position.

The selected IDs are returned chronologically by check-in date; equal dates retain input order. This deliberately differs from the apparent arrival order in one challenge example and makes the result operationally readable.

Operational times are fixed at:

- check-in: 15:00;
- checkout: 11:00 after the requested number of nights;
- cleaning: 11:00–12:00.

Therefore a booking may start on the calendar date when the previous booking checks out.

## Remaining doubts

None. These assumptions are explicit, tested, and described in the README and frontend FAQ so reviewers can challenge or replace them without reverse-engineering the implementation.
