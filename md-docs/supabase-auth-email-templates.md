# Supabase Auth Email Templates

Paste these templates into Supabase Dashboard -> Authentication -> Emails. They use the verified AutoSend sender `no-reply@mantrixflow.com` and keep replies/support text pointed at `support@mantrixflow.com`.

Supabase template variables used here are documented by Supabase: `{{ .ConfirmationURL }}`, `{{ .Token }}`, `{{ .SiteURL }}`, `{{ .RedirectTo }}`, `{{ .Email }}`, and `{{ .NewEmail }}`.

Important: do not paste this whole Markdown file into Supabase. For each Supabase template screen, copy only:

1. The subject text under `Subject`.
2. The raw HTML inside that template's `HTML` code block, starting at `<!DOCTYPE html>` and ending at `</html>`.

If the Supabase preview shows `# Supabase Auth Email Templates` or triple backticks, Markdown was pasted by mistake.

## Shared Notes

- Use AutoSend SMTP on port `465`.
- Set Supabase SMTP sender email to `no-reply@mantrixflow.com`.
- Keep the design minimal for deliverability: small logo, one CTA, fallback link/code, security copy.
- Do not enable click tracking on auth links.

## Confirm Sign Up

Subject:

```txt
Confirm your MantrixFlow account
```

HTML:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Confirm your MantrixFlow account</title>
  <style>
    body{margin:0;padding:0;background:#f7f7f7;-webkit-text-size-adjust:100%}table{border-collapse:collapse;border-spacing:0}img{border:0;display:block}a{color:#4f46e5;text-decoration:none}.wrap{width:100%;background:#f7f7f7}.shell{padding:48px 16px}.card{width:100%;max-width:640px;margin:0 auto;background:#fff;border:1px solid #e5e2df;border-radius:18px;overflow:hidden}.inner{padding:54px 48px 42px;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Arial,sans-serif;color:#111}.brand-row{margin:0 auto 46px;text-align:center}.brand-row img{display:inline-block;width:32px;height:auto;vertical-align:middle;margin:0 10px 0 0}.brand-row span{display:inline-block;vertical-align:middle;font-size:21px;line-height:1;font-weight:700;letter-spacing:.08em;text-transform:uppercase;color:#2b2b2b}.eyebrow{margin:0 0 18px;font-size:12px;font-weight:700;line-height:1.4;letter-spacing:.11em;text-transform:uppercase;color:#7a746f}h1{margin:0 0 22px;font-size:30px;line-height:1.25;font-weight:650;color:#050505}p{margin:0 0 20px;font-size:18px;line-height:1.55;color:#111}.code{margin:22px 0 24px;padding:18px 20px;background:#fafafa;border:1px solid #ebe8e5;border-radius:12px;font-size:28px;line-height:1.2;font-weight:700;letter-spacing:.08em;text-align:center}.cta{display:inline-block;margin:6px 0 18px;padding:13px 20px;border-radius:9px;background:#111;color:#fff!important;font-size:15px;font-weight:700;line-height:1}.fallback{margin:0 0 30px;font-size:13px;line-height:1.65;color:#817a75;word-break:break-all}.divider{height:1px;background:#e5e2df;margin:34px 0 24px}.footer p{margin:0 0 10px;font-size:15px;line-height:1.5;color:#77716d}.address{color:#9a9692!important}@media only screen and (max-width:640px){.shell{padding:18px 10px!important}.card{border-radius:16px!important}.inner{padding:34px 24px 30px!important}.brand-row{margin-bottom:34px!important}.brand-row img{width:28px!important}.brand-row span{font-size:18px!important}h1{font-size:25px!important}p{font-size:16px!important}.cta{display:block!important;text-align:center!important}.code{font-size:24px!important}}
  </style>
</head>
<body>
  <table role="presentation" width="100%" class="wrap"><tr><td align="center" class="shell">
    <table role="presentation" width="640" class="card"><tr><td class="inner">
      <div class="brand-row"><img src="https://d1v739xuxrzdgy.cloudfront.net/orgs/6a158503d8ca8942a075ce15/projects/6a158503d8ca8942a075ce18/media/1ce9920218c57a04ed.png" width="32" alt=""><span>MantrixFlow</span></div>
      <p class="eyebrow">Account confirmation</p>
      <h1>Confirm your email</h1>
      <p>Thanks for signing up for MantrixFlow. Confirm this email address to finish creating your account.</p>
      <a href="{{ .ConfirmationURL }}" class="cta">Confirm account</a>
      <p class="fallback">If the button does not work, open this link:<br><a href="{{ .ConfirmationURL }}">{{ .ConfirmationURL }}</a></p>
      <p class="fallback">Verification code: <strong>{{ .Token }}</strong></p>
      <div class="divider"></div>
      <div class="footer">
        <p>If you did not create a MantrixFlow account, you can safely ignore this email.</p>
        <p>Questions? Reach us at <a href="mailto:support@mantrixflow.com">support@mantrixflow.com</a></p>
        <p class="address">MantrixFlow</p>
      </div>
    </td></tr></table>
  </td></tr></table>
</body>
</html>
```

## Invite User

Subject:

```txt
You're invited to MantrixFlow
```

HTML:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>You're invited to MantrixFlow</title>
  <style>
    body{margin:0;padding:0;background:#f7f7f7;-webkit-text-size-adjust:100%}table{border-collapse:collapse;border-spacing:0}img{border:0;display:block}a{color:#4f46e5;text-decoration:none}.wrap{width:100%;background:#f7f7f7}.shell{padding:48px 16px}.card{width:100%;max-width:640px;margin:0 auto;background:#fff;border:1px solid #e5e2df;border-radius:18px;overflow:hidden}.inner{padding:54px 48px 42px;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Arial,sans-serif;color:#111}.brand-row{margin:0 auto 46px;text-align:center}.brand-row img{display:inline-block;width:32px;height:auto;vertical-align:middle;margin:0 10px 0 0}.brand-row span{display:inline-block;vertical-align:middle;font-size:21px;line-height:1;font-weight:700;letter-spacing:.08em;text-transform:uppercase;color:#2b2b2b}.eyebrow{margin:0 0 18px;font-size:12px;font-weight:700;line-height:1.4;letter-spacing:.11em;text-transform:uppercase;color:#7a746f}h1{margin:0 0 22px;font-size:30px;line-height:1.25;font-weight:650;color:#050505}p{margin:0 0 20px;font-size:18px;line-height:1.55;color:#111}.summary{width:100%;margin:30px 0 28px;border-top:1px solid #e5e2df;border-bottom:1px solid #e5e2df}.summary td{padding:14px 0;border-bottom:1px solid #efedeb;font-size:15px;line-height:1.45;vertical-align:top}.summary tr:last-child td{border-bottom:0}.summary td:first-child{width:38%;padding-right:18px;color:#817a75}.summary td:last-child{color:#111;font-weight:600;text-align:right}.cta{display:inline-block;margin:6px 0 18px;padding:13px 20px;border-radius:9px;background:#111;color:#fff!important;font-size:15px;font-weight:700;line-height:1}.fallback{margin:0 0 30px;font-size:13px;line-height:1.65;color:#817a75;word-break:break-all}.divider{height:1px;background:#e5e2df;margin:34px 0 24px}.footer p{margin:0 0 10px;font-size:15px;line-height:1.5;color:#77716d}.address{color:#9a9692!important}@media only screen and (max-width:640px){.shell{padding:18px 10px!important}.card{border-radius:16px!important}.inner{padding:34px 24px 30px!important}.brand-row{margin-bottom:34px!important}.brand-row img{width:28px!important}.brand-row span{font-size:18px!important}h1{font-size:25px!important}p{font-size:16px!important}.summary td{display:block!important;width:100%!important;padding:8px 0!important;text-align:left!important;border-bottom:0!important}.summary tr{display:block!important;padding:10px 0!important;border-bottom:1px solid #efedeb!important}.cta{display:block!important;text-align:center!important}}
  </style>
</head>
<body>
  <table role="presentation" width="100%" class="wrap"><tr><td align="center" class="shell">
    <table role="presentation" width="640" class="card"><tr><td class="inner">
      <div class="brand-row"><img src="https://d1v739xuxrzdgy.cloudfront.net/orgs/6a158503d8ca8942a075ce15/projects/6a158503d8ca8942a075ce18/media/1ce9920218c57a04ed.png" width="32" alt=""><span>MantrixFlow</span></div>
      <p class="eyebrow">Workspace invitation</p>
      <h1>Join MantrixFlow</h1>
      <p>You were invited to create a MantrixFlow account and collaborate in a workspace.</p>
      <table role="presentation" class="summary" width="100%">
        <tr><td>Email</td><td>{{ .Email }}</td></tr>
      </table>
      <a href="{{ .ConfirmationURL }}" class="cta">Accept invitation</a>
      <p class="fallback">If the button does not work, open this link:<br><a href="{{ .ConfirmationURL }}">{{ .ConfirmationURL }}</a></p>
      <div class="divider"></div>
      <div class="footer">
        <p>You received this because a workspace admin invited this email address to MantrixFlow.</p>
        <p>Questions? Reach us at <a href="mailto:support@mantrixflow.com">support@mantrixflow.com</a></p>
        <p class="address">MantrixFlow</p>
      </div>
    </td></tr></table>
  </td></tr></table>
</body>
</html>
```

## Magic Link Or OTP

Subject:

```txt
Your MantrixFlow sign-in link
```

HTML:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Your MantrixFlow sign-in link</title>
  <style>
    body{margin:0;padding:0;background:#f7f7f7;-webkit-text-size-adjust:100%}table{border-collapse:collapse;border-spacing:0}img{border:0;display:block}a{color:#4f46e5;text-decoration:none}.wrap{width:100%;background:#f7f7f7}.shell{padding:48px 16px}.card{width:100%;max-width:640px;margin:0 auto;background:#fff;border:1px solid #e5e2df;border-radius:18px;overflow:hidden}.inner{padding:54px 48px 42px;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Arial,sans-serif;color:#111}.brand-row{margin:0 auto 46px;text-align:center}.brand-row img{display:inline-block;width:32px;height:auto;vertical-align:middle;margin:0 10px 0 0}.brand-row span{display:inline-block;vertical-align:middle;font-size:21px;line-height:1;font-weight:700;letter-spacing:.08em;text-transform:uppercase;color:#2b2b2b}.eyebrow{margin:0 0 18px;font-size:12px;font-weight:700;line-height:1.4;letter-spacing:.11em;text-transform:uppercase;color:#7a746f}h1{margin:0 0 22px;font-size:30px;line-height:1.25;font-weight:650;color:#050505}p{margin:0 0 20px;font-size:18px;line-height:1.55;color:#111}.code{margin:22px 0 24px;padding:18px 20px;background:#fafafa;border:1px solid #ebe8e5;border-radius:12px;font-size:28px;line-height:1.2;font-weight:700;letter-spacing:.08em;text-align:center}.cta{display:inline-block;margin:6px 0 18px;padding:13px 20px;border-radius:9px;background:#111;color:#fff!important;font-size:15px;font-weight:700;line-height:1}.fallback{margin:0 0 30px;font-size:13px;line-height:1.65;color:#817a75;word-break:break-all}.divider{height:1px;background:#e5e2df;margin:34px 0 24px}.footer p{margin:0 0 10px;font-size:15px;line-height:1.5;color:#77716d}.address{color:#9a9692!important}@media only screen and (max-width:640px){.shell{padding:18px 10px!important}.card{border-radius:16px!important}.inner{padding:34px 24px 30px!important}.brand-row{margin-bottom:34px!important}.brand-row img{width:28px!important}.brand-row span{font-size:18px!important}h1{font-size:25px!important}p{font-size:16px!important}.cta{display:block!important;text-align:center!important}.code{font-size:24px!important}}
  </style>
</head>
<body>
  <table role="presentation" width="100%" class="wrap"><tr><td align="center" class="shell">
    <table role="presentation" width="640" class="card"><tr><td class="inner">
      <div class="brand-row"><img src="https://d1v739xuxrzdgy.cloudfront.net/orgs/6a158503d8ca8942a075ce15/projects/6a158503d8ca8942a075ce18/media/1ce9920218c57a04ed.png" width="32" alt=""><span>MantrixFlow</span></div>
      <p class="eyebrow">Sign in</p>
      <h1>Open MantrixFlow</h1>
      <p>Use this secure link to sign in. It expires shortly and can only be used once.</p>
      <a href="{{ .ConfirmationURL }}" class="cta">Sign in</a>
      <p class="fallback">If the button does not work, open this link:<br><a href="{{ .ConfirmationURL }}">{{ .ConfirmationURL }}</a></p>
      <p class="fallback">One-time code:</p>
      <div class="code">{{ .Token }}</div>
      <div class="divider"></div>
      <div class="footer">
        <p>If you did not request this sign-in email, you can safely ignore it.</p>
        <p>Questions? Reach us at <a href="mailto:support@mantrixflow.com">support@mantrixflow.com</a></p>
        <p class="address">MantrixFlow</p>
      </div>
    </td></tr></table>
  </td></tr></table>
</body>
</html>
```

## Change Email Address

Subject:

```txt
Confirm your new MantrixFlow email
```

HTML:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Confirm your new email address</title>
  <style>
    body{margin:0;padding:0;background:#f7f7f7;-webkit-text-size-adjust:100%}table{border-collapse:collapse;border-spacing:0}img{border:0;display:block}a{color:#4f46e5;text-decoration:none}.wrap{width:100%;background:#f7f7f7}.shell{padding:48px 16px}.card{width:100%;max-width:640px;margin:0 auto;background:#fff;border:1px solid #e5e2df;border-radius:18px;overflow:hidden}.inner{padding:54px 48px 42px;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Arial,sans-serif;color:#111}.brand-row{margin:0 auto 46px;text-align:center}.brand-row img{display:inline-block;width:32px;height:auto;vertical-align:middle;margin:0 10px 0 0}.brand-row span{display:inline-block;vertical-align:middle;font-size:21px;line-height:1;font-weight:700;letter-spacing:.08em;text-transform:uppercase;color:#2b2b2b}.eyebrow{margin:0 0 18px;font-size:12px;font-weight:700;line-height:1.4;letter-spacing:.11em;text-transform:uppercase;color:#7a746f}h1{margin:0 0 22px;font-size:30px;line-height:1.25;font-weight:650;color:#050505}p{margin:0 0 20px;font-size:18px;line-height:1.55;color:#111}.summary{width:100%;margin:30px 0 28px;border-top:1px solid #e5e2df;border-bottom:1px solid #e5e2df}.summary td{padding:14px 0;border-bottom:1px solid #efedeb;font-size:15px;line-height:1.45;vertical-align:top}.summary tr:last-child td{border-bottom:0}.summary td:first-child{width:38%;padding-right:18px;color:#817a75}.summary td:last-child{color:#111;font-weight:600;text-align:right}.cta{display:inline-block;margin:6px 0 18px;padding:13px 20px;border-radius:9px;background:#111;color:#fff!important;font-size:15px;font-weight:700;line-height:1}.fallback{margin:0 0 30px;font-size:13px;line-height:1.65;color:#817a75;word-break:break-all}.divider{height:1px;background:#e5e2df;margin:34px 0 24px}.footer p{margin:0 0 10px;font-size:15px;line-height:1.5;color:#77716d}.address{color:#9a9692!important}@media only screen and (max-width:640px){.shell{padding:18px 10px!important}.card{border-radius:16px!important}.inner{padding:34px 24px 30px!important}.brand-row{margin-bottom:34px!important}.brand-row img{width:28px!important}.brand-row span{font-size:18px!important}h1{font-size:25px!important}p{font-size:16px!important}.summary td{display:block!important;width:100%!important;padding:8px 0!important;text-align:left!important;border-bottom:0!important}.summary tr{display:block!important;padding:10px 0!important;border-bottom:1px solid #efedeb!important}.cta{display:block!important;text-align:center!important}}
  </style>
</head>
<body>
  <table role="presentation" width="100%" class="wrap"><tr><td align="center" class="shell">
    <table role="presentation" width="640" class="card"><tr><td class="inner">
      <div class="brand-row"><img src="https://d1v739xuxrzdgy.cloudfront.net/orgs/6a158503d8ca8942a075ce15/projects/6a158503d8ca8942a075ce18/media/1ce9920218c57a04ed.png" width="32" alt=""><span>MantrixFlow</span></div>
      <p class="eyebrow">Email change</p>
      <h1>Confirm your new email</h1>
      <p>Confirm that you want to use this email address for your MantrixFlow account.</p>
      <table role="presentation" class="summary" width="100%">
        <tr><td>Current email</td><td>{{ .Email }}</td></tr>
        <tr><td>New email</td><td>{{ .NewEmail }}</td></tr>
      </table>
      <a href="{{ .ConfirmationURL }}" class="cta">Confirm new email</a>
      <p class="fallback">If the button does not work, open this link:<br><a href="{{ .ConfirmationURL }}">{{ .ConfirmationURL }}</a></p>
      <div class="divider"></div>
      <div class="footer">
        <p>If you did not request this change, contact support immediately.</p>
        <p>Questions? Reach us at <a href="mailto:support@mantrixflow.com">support@mantrixflow.com</a></p>
        <p class="address">MantrixFlow</p>
      </div>
    </td></tr></table>
  </td></tr></table>
</body>
</html>
```

## Reset Password

Subject:

```txt
Reset your MantrixFlow password
```

HTML:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Reset your MantrixFlow password</title>
  <style>
    body{margin:0;padding:0;background:#f7f7f7;-webkit-text-size-adjust:100%}table{border-collapse:collapse;border-spacing:0}img{border:0;display:block}a{color:#4f46e5;text-decoration:none}.wrap{width:100%;background:#f7f7f7}.shell{padding:48px 16px}.card{width:100%;max-width:640px;margin:0 auto;background:#fff;border:1px solid #e5e2df;border-radius:18px;overflow:hidden}.inner{padding:54px 48px 42px;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Arial,sans-serif;color:#111}.brand-row{margin:0 auto 46px;text-align:center}.brand-row img{display:inline-block;width:32px;height:auto;vertical-align:middle;margin:0 10px 0 0}.brand-row span{display:inline-block;vertical-align:middle;font-size:21px;line-height:1;font-weight:700;letter-spacing:.08em;text-transform:uppercase;color:#2b2b2b}.eyebrow{margin:0 0 18px;font-size:12px;font-weight:700;line-height:1.4;letter-spacing:.11em;text-transform:uppercase;color:#7a746f}h1{margin:0 0 22px;font-size:30px;line-height:1.25;font-weight:650;color:#050505}p{margin:0 0 20px;font-size:18px;line-height:1.55;color:#111}.cta{display:inline-block;margin:6px 0 18px;padding:13px 20px;border-radius:9px;background:#111;color:#fff!important;font-size:15px;font-weight:700;line-height:1}.fallback{margin:0 0 30px;font-size:13px;line-height:1.65;color:#817a75;word-break:break-all}.divider{height:1px;background:#e5e2df;margin:34px 0 24px}.footer p{margin:0 0 10px;font-size:15px;line-height:1.5;color:#77716d}.address{color:#9a9692!important}@media only screen and (max-width:640px){.shell{padding:18px 10px!important}.card{border-radius:16px!important}.inner{padding:34px 24px 30px!important}.brand-row{margin-bottom:34px!important}.brand-row img{width:28px!important}.brand-row span{font-size:18px!important}h1{font-size:25px!important}p{font-size:16px!important}.cta{display:block!important;text-align:center!important}}
  </style>
</head>
<body>
  <table role="presentation" width="100%" class="wrap"><tr><td align="center" class="shell">
    <table role="presentation" width="640" class="card"><tr><td class="inner">
      <div class="brand-row"><img src="https://d1v739xuxrzdgy.cloudfront.net/orgs/6a158503d8ca8942a075ce15/projects/6a158503d8ca8942a075ce18/media/1ce9920218c57a04ed.png" width="32" alt=""><span>MantrixFlow</span></div>
      <p class="eyebrow">Password reset</p>
      <h1>Reset your password</h1>
      <p>We received a request to reset the password for your MantrixFlow account.</p>
      <a href="{{ .ConfirmationURL }}" class="cta">Reset password</a>
      <p class="fallback">If the button does not work, open this link:<br><a href="{{ .ConfirmationURL }}">{{ .ConfirmationURL }}</a></p>
      <div class="divider"></div>
      <div class="footer">
        <p>If you did not request a password reset, you can safely ignore this email.</p>
        <p>Questions? Reach us at <a href="mailto:support@mantrixflow.com">support@mantrixflow.com</a></p>
        <p class="address">MantrixFlow</p>
      </div>
    </td></tr></table>
  </td></tr></table>
</body>
</html>
```

## Reauthentication

Subject:

```txt
{{ .Token }} is your MantrixFlow verification code
```

HTML:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Your MantrixFlow verification code</title>
  <style>
    body{margin:0;padding:0;background:#f7f7f7;-webkit-text-size-adjust:100%}table{border-collapse:collapse;border-spacing:0}img{border:0;display:block}a{color:#4f46e5;text-decoration:none}.wrap{width:100%;background:#f7f7f7}.shell{padding:48px 16px}.card{width:100%;max-width:640px;margin:0 auto;background:#fff;border:1px solid #e5e2df;border-radius:18px;overflow:hidden}.inner{padding:54px 48px 42px;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Arial,sans-serif;color:#111}.brand-row{margin:0 auto 46px;text-align:center}.brand-row img{display:inline-block;width:32px;height:auto;vertical-align:middle;margin:0 10px 0 0}.brand-row span{display:inline-block;vertical-align:middle;font-size:21px;line-height:1;font-weight:700;letter-spacing:.08em;text-transform:uppercase;color:#2b2b2b}.eyebrow{margin:0 0 18px;font-size:12px;font-weight:700;line-height:1.4;letter-spacing:.11em;text-transform:uppercase;color:#7a746f}h1{margin:0 0 22px;font-size:30px;line-height:1.25;font-weight:650;color:#050505}p{margin:0 0 20px;font-size:18px;line-height:1.55;color:#111}.code{margin:22px 0 24px;padding:18px 20px;background:#fafafa;border:1px solid #ebe8e5;border-radius:12px;font-size:32px;line-height:1.2;font-weight:700;letter-spacing:.1em;text-align:center}.divider{height:1px;background:#e5e2df;margin:34px 0 24px}.footer p{margin:0 0 10px;font-size:15px;line-height:1.5;color:#77716d}.address{color:#9a9692!important}@media only screen and (max-width:640px){.shell{padding:18px 10px!important}.card{border-radius:16px!important}.inner{padding:34px 24px 30px!important}.brand-row{margin-bottom:34px!important}.brand-row img{width:28px!important}.brand-row span{font-size:18px!important}h1{font-size:25px!important}p{font-size:16px!important}.code{font-size:26px!important}}
  </style>
</head>
<body>
  <table role="presentation" width="100%" class="wrap"><tr><td align="center" class="shell">
    <table role="presentation" width="640" class="card"><tr><td class="inner">
      <div class="brand-row"><img src="https://d1v739xuxrzdgy.cloudfront.net/orgs/6a158503d8ca8942a075ce15/projects/6a158503d8ca8942a075ce18/media/1ce9920218c57a04ed.png" width="32" alt=""><span>MantrixFlow</span></div>
      <p class="eyebrow">Security verification</p>
      <h1>Verify your identity</h1>
      <p>Use this code to continue your sensitive MantrixFlow action. It expires shortly.</p>
      <div class="code">{{ .Token }}</div>
      <div class="divider"></div>
      <div class="footer">
        <p>If you did not request this code, secure your account and contact support.</p>
        <p>Questions? Reach us at <a href="mailto:support@mantrixflow.com">support@mantrixflow.com</a></p>
        <p class="address">MantrixFlow</p>
      </div>
    </td></tr></table>
  </td></tr></table>
</body>
</html>
```
