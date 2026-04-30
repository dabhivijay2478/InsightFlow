---
name: seat-based-billing
description: Guide for implementing Dodo Payments seat-based billing with add-ons, per-seat pricing, proration, subscription changes, webhooks, and application-side seat enforcement.
---

# Dodo Payments Seat-Based Billing

Reference: https://docs.dodopayments.com/features/seat-based-billing

Seat-based billing charges customers by the number of users, members, licenses, hosts, editors, or active seats in a workspace. Dodo Payments models seat billing through subscription products plus add-ons.

## Core Model

Use a base subscription product for the plan and an add-on for extra seats.

Example:

- Base product: Team Pro, USD 99/month, includes 5 seats.
- Seat add-on: Extra Seat, USD 15/month each.
- Customer needs 15 total seats.
- Billable add-on quantity: 10 extra seats.
- Monthly total: USD 99 + (10 * USD 15) = USD 249.

## Setup Flow

1. Decide the included seat count for each base plan.
2. Create an add-on in Dodo for additional seats.
3. Attach the add-on to the subscription product.
4. Start checkout with the base product and selected add-on quantity.
5. Store the Dodo subscription ID, base included seats, add-on ID, purchased add-on quantity, and total seat allowance.
6. Update local seat allowance from verified webhooks.
7. Enforce seat limits inside the application.

## Pricing Patterns

Base plus per-seat add-on:

- Best when small teams can use the base product.
- Example: USD 49/month includes 3 seats, then USD 10 per extra seat.

Pure per-seat pricing:

- Best for simple pricing.
- Use a zero-price or low-base subscription and charge every seat through the add-on.

Tiered seat pricing:

- Best when higher plans have cheaper seats or more features.
- Create separate products with different add-on pricing.

Seat bundles:

- Best when you want simpler purchasing and larger commitments.
- Create add-ons for 5-seat, 10-seat, or 25-seat packs.

## Changing Seats

When changing seat counts mid-cycle, use Dodo subscription plan-change APIs and pass the updated add-on quantities. Preview the change when possible before applying it.

Common proration modes:

- `prorated_immediately`: charge or credit based on remaining days in the cycle.
- `difference_immediately`: charge or credit the full difference immediately.
- `full_immediately`: charge the full new amount and reset the billing cycle.

Application flow:

1. Calculate current used seats.
2. Validate the requested new seat allowance is not below used seats unless the app first removes users.
3. Preview the billing change.
4. Confirm with the workspace admin.
5. Apply the subscription change in Dodo.
6. Wait for webhook confirmation or reconcile the subscription before updating durable access.

## Webhooks

Handle these subscription events for seat state:

- `subscription.active`: provision the initial seat allowance.
- `subscription.plan_changed`: update seat allowance after add-ons change.
- `subscription.renewed`: confirm the subscription and seat allowance are still valid.
- `subscription.cancelled`: deprovision paid seats or schedule access removal.
- `subscription.expired`: remove paid access after the term ends.

Webhook processing must be idempotent. Store processed event IDs and make local updates in a transaction.

## Seat Enforcement

Dodo tracks billing. Your application must enforce who can use the seats.

Hard limit:

- Block invitations when `used_seats >= total_seats`.
- Good for clear compliance and simple accounting.

Soft limit:

- Allow temporary overage, notify admins, and reconcile billing later.
- Good for sales-assisted or enterprise workflows.

Auto-upgrade:

- When a user is invited beyond the current allowance, add another seat through Dodo and notify the admin.
- Good for frictionless team growth, but use clear admin consent and audit logs.

## Data To Store Locally

- Dodo customer ID.
- Dodo subscription ID.
- Dodo base product ID.
- Seat add-on ID.
- Base included seats.
- Paid add-on seat quantity.
- Total allowed seats.
- Used seats.
- Billing status.
- Last processed Dodo event ID and timestamp.

## Review Checklist

- The server validates seat changes before calling Dodo.
- The UI previews immediate charge or credit before applying a seat change.
- The app does not trust client-submitted seat counts.
- Webhooks are signature-verified and idempotent.
- Seat limits are enforced in invite, activation, and role-change paths.
- Downgrades cannot strand active users in an undefined access state.
