# Start your first website security check

You can request a check without security expertise. You will need permission to test the website and access to its settings, hosting or DNS. Your website manager can help with verification and repairs.

## Before you buy

The standard service checks the public website without logging in. It looks for the external exposures and configuration issues listed in the [product scope](/product-scope). It does not test signed-in features, access between accounts or business workflows.

1 credit pays for 1 check of 1 hostname, including a repeat check. Different subdomains are separate targets. Check the [current prices](/#pricing) before purchasing. Adding a website and verifying control are free.

## Add your website

[Sign in](/login) with Google, open My websites and enter the address. Use the exact hostname you want checked: `www.example.com` and `api.example.com` are separate targets.

## Confirm access

Choose one of these methods on the website's card:

- **Website settings:** copy the generated meta tag into the head of the public homepage. This suits platforms that let you edit that setting.
- **Hosting file:** download the generated verification file and serve it at `/.well-known/netsekurity-verification.txt`. An API can serve the same content at that endpoint.
- **Domain settings:** add the generated DNS TXT record. Use the exact record name shown, accounting for any domain suffix your DNS provider adds automatically.

File and homepage verification require valid HTTPS and a direct response without a redirect or login. A website that redirects to another hostname can use DNS verification or verify the final hostname instead.

Press the check button for the method you used. Leave the proof in place: Netsekurity checks it again before each self-service scan. If it cannot be confirmed, the scan does not start and no credit is used. Review the instructions and try again; DNS changes may take time to appear.

**Someone else manages your website?** Open that option on the card and copy the instructions for them. Ask which method they used, then check it yourself. Do not send passwords or your Netsekurity login.

## Start the check

Buy credits when you are ready. Review the external, read-only scope and confirm that you are authorised to test this hostname. Each scan takes 10 minutes to 24 hours, depending on what it finds. One scan runs at a time per account. You can leave the dashboard while it runs.

## Read the result and plan repairs

Open Security scans. The result summary explains the next action in English and Indonesian; the downloadable report is a PDF in English. Give the report securely to your developer for evidence and repair guidance. Download links require the owning account or an authorised operator; forwarding the dashboard URL does not grant access.

The standard service provides findings and guidance. Your developer makes the repairs. After changes, start another check for 1 credit and review the new report.

No reported findings does not mean every part of the application is safe. A failed check is not a clean result; self-service failed scans receive a credit refund. If the result counts are unavailable, read the report rather than treating the missing counts as zero.

[View example results](/sample-report), [discuss comprehensive testing](/enterprise), or [read developer integration instructions](/docs).
