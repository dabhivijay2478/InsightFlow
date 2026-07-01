# AutoSend Template Design Guide

Use one shared MantrixFlow visual system for AutoSend templates. The style should feel like a clean verification email: compact brand row, a quiet white card, generous spacing, minimal decoration, one clear action, and a plain fallback link. Auth templates should use the plainest version for deliverability.

## Layout Rules

- Outer wrapper: `width: 100%; background: #f7f7f7; padding: 48px 16px;`.
- Email body: `max-width: 640px; width: 100%; margin: 0 auto;`.
- Use web-safe fonts: `Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", Arial, sans-serif`.
- Avoid absolute positioning, background images, scripts, forms, and external CSS.
- Inline critical CSS. A small `<style>` block for mobile rules is acceptable.
- Keep body copy to 2-4 short paragraphs.
- Use one primary CTA and one plain fallback URL.
- Use a small uppercase category label instead of loud colored badges.
- Footer must explain why the recipient received the email. Include preferences link for non-critical email.

## Color Tokens

| Token | Color | Use |
| --- | --- | --- |
| Ink | `#111111` | Main text |
| Muted | `#7a746f` | Secondary text |
| Border | `#e5e2df` | Card border and dividers |
| Background | `#f7f7f7` | Page background |
| Panel | `#ffffff` | Main card |
| Primary | `#111111` | CTA button |
| Link | `#4f46e5` | Fallback/support links |

## Responsive Behavior

```html
<style>
  @media only screen and (max-width: 640px) {
    .mf-wrap { padding: 18px 10px !important; }
    .mf-card { border-radius: 16px !important; }
    .mf-px { padding-left: 24px !important; padding-right: 24px !important; }
    .mf-button { display: block !important; width: 100% !important; box-sizing: border-box !important; }
    .mf-row-label, .mf-row-value { display: block !important; width: 100% !important; text-align: left !important; }
  }
</style>
```

## Base Shell

```html
<div class="mf-wrap" style="width:100%;background:#f7f7f7;padding:48px 16px;">
  <div class="mf-card" style="max-width:640px;width:100%;margin:0 auto;background:#ffffff;border:1px solid #e5e2df;border-radius:18px;overflow:hidden;font-family:Inter,-apple-system,BlinkMacSystemFont,'Segoe UI',Arial,sans-serif;color:#111111;">
    <div class="mf-px" style="padding:54px 48px 42px;">
      <div style="margin:0 auto 46px;text-align:center;">
        <img src="https://d1v739xuxrzdgy.cloudfront.net/orgs/6a158503d8ca8942a075ce15/projects/6a158503d8ca8942a075ce18/media/1ce9920218c57a04ed.png" width="32" alt="" style="display:inline-block;width:32px;height:auto;vertical-align:middle;margin:0 10px 0 0;">
        <span style="display:inline-block;vertical-align:middle;font-size:21px;line-height:1;font-weight:700;letter-spacing:.08em;text-transform:uppercase;color:#2b2b2b;">MantrixFlow</span>
      </div>
      <div style="font-size:12px;color:#7a746f;text-transform:uppercase;letter-spacing:.11em;font-weight:700;margin:0 0 18px;">{{category_label}}</div>
      <h1 style="font-size:30px;line-height:1.25;margin:0 0 22px;font-weight:650;color:#050505;">{{title}}</h1>
      <p style="font-size:18px;line-height:1.55;margin:0 0 20px;color:#111111;">{{body_intro}}</p>
    </div>

    <div class="mf-px" style="padding:0 48px 42px;">
      <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="border-collapse:collapse;border-top:1px solid #e5e2df;border-bottom:1px solid #e5e2df;margin:10px 0 28px;">
        <tr>
          <td class="mf-row-label" style="padding:14px 0;border-bottom:1px solid #efedeb;color:#817a75;font-size:15px;">Pipeline</td>
          <td class="mf-row-value" style="padding:14px 0;border-bottom:1px solid #efedeb;text-align:right;font-size:15px;font-weight:700;">{{pipeline_name}}</td>
        </tr>
      </table>
      <a class="mf-button" href="{{cta_url}}" style="display:inline-block;background:#111111;color:#ffffff;text-decoration:none;border-radius:9px;padding:13px 20px;font-size:15px;font-weight:700;">{{cta_label}}</a>
      <p style="font-size:13px;line-height:1.65;color:#817a75;margin:18px 0 30px;word-break:break-all;">If the button does not work, open this link:<br><a href="{{cta_url}}" style="color:#4f46e5;">{{cta_url}}</a></p>
      <div style="height:1px;background:#e5e2df;margin:34px 0 24px;"></div>
      <p style="font-size:15px;line-height:1.5;color:#77716d;margin:0 0 10px;">You received this because {{why_received}}.</p>
      <p style="font-size:15px;line-height:1.5;color:#77716d;margin:0;">Questions? Reach us at <a href="mailto:support@mantrixflow.com" style="color:#4f46e5;">support@mantrixflow.com</a></p>
    </div>
  </div>
</div>
```

## Template Checklist

- Subject is specific and under about 70 characters.
- Preview text adds context without repeating the subject.
- CTA points to `APP_WEB_URL`, not localhost.
- Auth/security/billing-critical email has no unsubscribe language.
- Digest, onboarding, and re-engagement include preferences text.
- Error emails show safe summaries only. Do not include secrets, DSNs, tokens, SQL passwords, or raw connector configs.
