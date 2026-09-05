# AWS SES Setup For MantrixFlow

This guide follows the current Amazon SES console flow for creating the SES
identity, DNS records, production access, and Supabase SMTP setup.

If the SES page shows this error:

```text
The resource you attempted to access doesn't exist.
Email identity mantrixflow.com does not exist.
```

that means the SES identity has not been created in the selected AWS region yet.
Go back to the SES **Identities** list and create it first.

## Target Setup

| Setting | Value |
| --- | --- |
| AWS region | `ap-south-1` / Asia Pacific (Mumbai) |
| SES domain identity | `mantrixflow.com` |
| Go API sender | `MantrixFlow <support@mantrixflow.com>` |
| Supabase Auth sender | `no-reply@mantrixflow.com` |
| Custom MAIL FROM domain | `mail.mantrixflow.com` |
| SES configuration set | `mantrixflow-transactional` |
| Supabase SMTP host | `email-smtp.ap-south-1.amazonaws.com` |
| Supabase SMTP port | `587` with STARTTLS |

MantrixFlow uses SES in two lanes:

- Go API product emails: rendered by Go from committed templates and sent with the SES API.
- Supabase Auth emails: signup, password reset, magic links, and auth links sent through SES SMTP.

## 1. Open The Correct SES Region

1. Open the AWS Console.
2. In the region picker, choose **Asia Pacific (Mumbai) ap-south-1**.
3. Search for **Amazon SES** and open it.
4. In the left navigation, under **Configuration**, choose **Identities**.

You should see the same page as the screenshot: an **Identities** table and a
**Create identity** button. If the table says `Identities (0)`, that is fine.

Do not keep refreshing the missing identity detail URL. It cannot work until the
identity is created in this region.

## 2. Create The Domain Identity

From **SES > Configuration > Identities**:

1. Click **Create identity**.
2. In **Identity details**, select **Domain**.
3. In **Domain**, enter:

```text
mantrixflow.com
```

4. Do not enter `www.mantrixflow.com`.
5. If **Assign a default configuration set** is visible:
   - Leave it unchecked if `mantrixflow-transactional` does not exist yet.
   - If the configuration set already exists, check it and select `mantrixflow-transactional`.
6. If **Use a custom MAIL FROM domain** is visible in this create form:
   - Check it.
   - MAIL FROM domain: `mail.mantrixflow.com`.
   - Behavior on MX failure: choose **Use default MAIL FROM domain** while DNS is still propagating.
7. Under **Verifying your domain**, keep the default **Easy DKIM** setup.
8. If **Advanced DKIM settings** is expanded or visible:
   - Identity type: **Easy DKIM**.
   - DKIM signing key length: **RSA_2048_BIT**.
9. Ensure **DKIM signatures** is enabled.
10. Tags are optional.
11. Click **Create identity**.

After creation, SES opens the identity detail page for `mantrixflow.com`. The
identity will remain unverified until DNS records are published.

Domain identity verification is enough to send from addresses under
`mantrixflow.com`, including `support@mantrixflow.com` and
`no-reply@mantrixflow.com`. Create separate email address identities only if you
need per-address custom settings later.

## 3. Copy The DKIM DNS Records From SES

Open the `mantrixflow.com` identity:

1. Go to **SES > Configuration > Identities**.
2. Click the `mantrixflow.com` row.
3. Open the **Authentication** tab.
4. Find **Publish DNS records**.
5. Copy all three Easy DKIM **CNAME** records.

In Cloudflare:

1. Open the `mantrixflow.com` zone.
2. Go to **DNS > Records**.
3. Add each SES DKIM CNAME record exactly as shown.
4. Set each record to **DNS only**.
5. Do not proxy SES CNAME, MX, or TXT records.

Cloudflare entry rules:

- If SES shows a name ending in `.mantrixflow.com`, Cloudflare may only need the host portion before the root domain.
- Do not duplicate the domain, for example do not create `selector._domainkey.mantrixflow.com.mantrixflow.com`.
- Do not remove the required underscore in `_domainkey`.
- Copy the target/value exactly from SES.

Wait for SES to show:

```text
Identity status: Verified
DKIM status: Successful
```

DNS propagation can take up to 72 hours, but Cloudflare usually updates faster.

## 4. Configure Custom MAIL FROM

If custom MAIL FROM was configured during identity creation, stay on the identity
detail page and copy the MAIL FROM DNS records from the **Custom MAIL FROM
domain** area.

If it was not configured during creation:

1. Go to **SES > Configuration > Identities**.
2. Click `mantrixflow.com`.
3. On the identity detail page, find **Custom MAIL FROM domain**.
4. Click **Edit**.
5. Check **Use a custom MAIL FROM domain**.
6. MAIL FROM domain:

```text
mail.mantrixflow.com
```

7. Behavior on MX failure:
   - During setup, choose **Use default MAIL FROM domain**.
   - After SES shows MAIL FROM as successful, you can edit this to **Reject message** if you want strict failure behavior.
8. Click **Save changes**.

SES will show an MX record and an SPF TXT record. Add them in Cloudflare.

For `ap-south-1`, they normally look like this:

| Type | Cloudflare name | Value |
| --- | --- | --- |
| MX | `mail` | `10 feedback-smtp.ap-south-1.amazonses.com` |
| TXT | `mail` | `v=spf1 include:amazonses.com ~all` |

Important:

- The MAIL FROM domain must have exactly one MX record for SES.
- Keep these records **DNS only** in Cloudflare.
- If Cloudflare has separate priority and mail-server fields for MX, set priority to `10` and server to `feedback-smtp.ap-south-1.amazonses.com`.

Wait until SES shows the custom MAIL FROM status as successful.

## 5. Add Or Confirm DMARC

In Cloudflare, add a DMARC TXT record for the root domain if it does not already
exist:

```text
Type: TXT
Name: _dmarc
Value: v=DMARC1; p=none; rua=mailto:dmarc@mantrixflow.com; adkim=s; aspf=s
```

Start with `p=none` while validating delivery. Move to `quarantine` or `reject`
only after production email consistently passes SPF, DKIM, and DMARC.

## 6. Create The SES Configuration Set

In the SES console:

1. In the left navigation, under **Configuration**, choose **Configuration sets**.
2. Click **Create set**.
3. Configuration set name:

```text
mantrixflow-transactional
```

4. Keep sending enabled.
5. Create the set.

The Go API sends this configuration set name in SES API calls.

### Add The SES Event Destination

Create an SNS topic first:

1. Open **Amazon SNS** in `ap-south-1`.
2. Choose **Topics**.
3. Click **Create topic**.
4. Type: **Standard**.
5. Name:

```text
mantrixflow-ses-events
```

6. Create the topic.
7. Create a subscription:
   - Protocol: **HTTPS**
   - Endpoint: `https://<api-domain>/api/v1/webhooks/ses`

Then return to SES:

1. Open **SES > Configuration > Configuration sets**.
2. Open `mantrixflow-transactional`.
3. Open **Event destinations**.
4. Click **Add destination**.
5. Destination type: **Amazon SNS**.
6. SNS topic: `mantrixflow-ses-events`.
7. Select these event types:
   - Sends
   - Deliveries
   - Bounces
   - Complaints
   - Rejects
   - Delivery delays
8. Save the destination.

The Go API verifies AWS SNS signatures, handles the subscription confirmation,
stores provider events, updates `email_jobs.delivery_status`, and suppresses
future sends after permanent bounces or complaints.

## 7. Request Production Access

New SES accounts are usually in sandbox mode. Sandbox mode can send only to
verified recipients, so production email will not work until production access is
approved.

In SES:

1. Open **Account dashboard**.
2. Find the sandbox / production access panel.
3. Click **Request production access**.
4. Mail type: **Transactional**.
5. Website URL:

```text
https://mantrixflow.com
```

6. Use case:

```text
MantrixFlow sends transactional product emails: organization invites,
pipeline alerts, billing notices, onboarding emails, weekly digests, signup
emails, magic links, and password reset emails. Marketing/campaign email is
out of scope.
```

7. Add operational notes:

```text
The Go API records delivery jobs in email_jobs, stores SES message IDs, retries
transient failures, and skips invalid/disposable recipients. Supabase Auth uses
SES SMTP for auth links. SES events from the mantrixflow-transactional
configuration set are delivered to the Go API through SNS, where bounces and
complaints create recipient suppressions.
```

8. Submit the request.

Do not run production sends to unverified recipients until AWS approves
production access.

## 8. Configure Go API Product Email

The Go API uses the ECS task role for SES API sending. The role must allow:

```text
ses:SendEmail
ses:SendRawEmail
```

Production environment:

```bash
EMAIL_PROVIDER=ses
EMAIL_FROM="MantrixFlow <support@mantrixflow.com>"
EMAIL_LOGO_URL=https://cloud.mantrixflow.com/m.png
AWS_SES_REGION=ap-south-1
AWS_SES_CONFIGURATION_SET=mantrixflow-transactional
AWS_SES_EVENT_SNS_TOPIC_ARN=arn:aws:sns:ap-south-1:<account-id>:mantrixflow-ses-events
```

Local dry run:

```bash
cd apps/server/arcyria-server
go run ./cmd/emailtest -dry-run
```

Live staging send:

```bash
cd apps/server/arcyria-server
TEST_EMAIL=you@example.com go run ./cmd/emailtest
```

In sandbox mode, the `TEST_EMAIL` recipient must also be verified in SES.

## 9. Configure Supabase Auth SMTP With SES

Supabase Auth emails do not use the Go queue. Configure Supabase custom SMTP so
signup, reset, invite, and magic-link email uses SES.

Create SES SMTP credentials:

1. Open SES in `ap-south-1`.
2. In the left navigation, choose **SMTP settings**.
3. Click **Create SMTP credentials**.
4. Create the IAM SMTP user.
5. Copy the SMTP username and password immediately.
6. Store them securely.

SES SMTP credentials are region-specific. They are not the same as normal AWS
access keys.

In Supabase Dashboard:

1. Open the project.
2. Go to **Authentication**.
3. Open **Emails** / **SMTP settings**.
4. Enable custom SMTP.
5. Configure:

```text
Host: email-smtp.ap-south-1.amazonaws.com
Port: 587
Username: SES SMTP username
Password: SES SMTP password
Sender email: no-reply@mantrixflow.com
Sender name: MantrixFlow
Secure connection: STARTTLS / TLS enabled
```

Do not enable link tracking or URL rewriting for auth emails. Rewritten magic
links and password reset links can break authentication.

## 10. Smoke Test

Run this in staging first.

Go API product emails:

- Organization invite.
- Pipeline failed alert.
- Billing/payment email.
- Weekly digest.

Supabase Auth emails:

- Signup email.
- Password reset email.
- Magic link email, if enabled.

For each delivered email:

- From address is correct.
- Links open correctly.
- DKIM passes for `mantrixflow.com`.
- SPF passes for `mail.mantrixflow.com` / the SES MAIL FROM domain.
- DMARC passes.
- SES message ID is stored for Go product emails.

## Troubleshooting

### The SES page says the identity does not exist

You are on the identity detail URL before creating the identity, or you are in
the wrong AWS region. Go to **SES > Configuration > Identities** in
`ap-south-1` and click **Create identity**.

### Identities table is empty

That is the expected first-time state. Click **Create identity** and create the
`mantrixflow.com` domain identity.

### Identity stays unverified

Open `mantrixflow.com` > **Authentication** > **Publish DNS records** and compare
the three DKIM CNAME records with Cloudflare. Check for duplicated domain names,
missing underscores, proxied records, or wrong region records.

### `MailFromDomainNotVerified`

Check the `mail.mantrixflow.com` MX and TXT records. There must be exactly one MX
record for the MAIL FROM domain, and it must point to
`feedback-smtp.ap-south-1.amazonses.com`.

### SES sends only to verified recipients

The account is still in sandbox mode. Request production access, or verify the
test recipient identity for sandbox testing.

### Supabase SMTP authentication fails

Regenerate SES SMTP credentials from **SMTP settings** in `ap-south-1`. Do not
use normal AWS IAM access keys as SMTP credentials.

### Supabase auth links are broken

Disable click tracking, link tracking, or any URL rewriting for Supabase Auth
emails.

## References

- [AWS SES create and verify identities](https://docs.aws.amazon.com/ses/latest/dg/creating-identities.html)
- [AWS SES custom MAIL FROM](https://docs.aws.amazon.com/ses/latest/dg/mail-from.html)
- [AWS SES configuration sets](https://docs.aws.amazon.com/ses/latest/dg/creating-configuration-sets.html)
- [AWS SES production access](https://docs.aws.amazon.com/ses/latest/dg/request-production-access.html)
- [AWS SES SMTP credentials](https://docs.aws.amazon.com/ses/latest/dg/smtp-credentials.html)
- [Supabase custom SMTP](https://supabase.com/docs/guides/auth/auth-smtp)
