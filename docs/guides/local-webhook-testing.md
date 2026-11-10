---
title: Local Webhook Testing: Stripe, GitHub, Shopify & Slack
description: Step-by-step guide to testing incoming webhooks on localhost using Expose. Learn how to verify HMAC signatures, handle HTTPS callbacks, and preserve stable subdomains without paid plans.
keywords:
  - local webhook testing
  - test stripe webhooks locally
  - github webhook localhost
  - shopify webhook localhost
  - ngrok alternative webhook
  - webhook testing tool
  - localhost tunnel
---

# Local Webhook Testing with Expose

When building integrations with services like **Stripe**, **GitHub**, **Shopify**, **Slack**, or **Twilio**, your application must handle asynchronous HTTP POST webhook events (such as `payment_intent.succeeded`, `push`, or `orders/create`).

Because external webhook providers cannot dispatch HTTP requests to `http://localhost:3000`, developers need a reliable, publicly reachable HTTPS tunnel.

**Expose** provides an open-source, single-binary solution for webhook testing with zero signups, zero configuration, and strict byte-for-byte signature integrity.

---

## The Webhook Architecture with Expose

```text
[ Webhook Provider (e.g., Stripe, GitHub) ]
       │ Dispatches HTTPS POST with HMAC signature
       ▼
[ Public Tunnel URL (e.g., https://swift-falcon-4821.loca.lt) ]
       │
[ Expose Relay Gateway ]
       │ Layer 4 Raw Socket Splicing (zero payload mutation)
       ▼
[ Local Expose Client ]
       │ Forwards byte-for-byte
       ▼
[ Local Backend Application (http://localhost:3000/api/webhooks) ]
       └── Computes HMAC over raw body ──► Signature Verified ✅
```

---

## Why Expose Excels at Webhook Testing

### 1. Byte-for-Byte Signature Integrity (HMAC-SHA256)
Security-critical webhooks (like Stripe's `Stripe-Signature` header or GitHub's `X-Hub-Signature-256`) compute a cryptographic HMAC signature over the exact binary bytes of the incoming request body. 

Some HTTP reverse proxies normalize whitespace, modify header casing, or strip trailing bytes during re-marshaling, which breaks signature verification. Expose performs **Layer 4 raw socket splicing** (`io.CopyBuffer`), guaranteeing that every byte and header sent by the webhook provider arrives at your local server unaltered.

### 2. Auto-Reconnecting Port Watcher
If you start your tunnel before your local development server is running, Expose detects this automatically:

```text
🚀 Tunnel[localtunnel] started for localhost:3000
✓ Public URL: https://swift-falcon-4821.loca.lt
✓ Forwarding to: http://localhost:3000 (Waiting for local app to start...)
```

Once your dev server boots up (e.g. Next.js, Rails, Django, Go), Expose detects the active socket and begins routing webhook events immediately—without restarting the tunnel.

---

## Quickstart: Testing Stripe Webhooks

Here is a complete walkthrough for testing Stripe payment webhooks locally.

### Step 1: Start your Local Webhook Handler

Below is a standard Node.js Express webhook listener:

```javascript
// server.js
const express = require('express');
const stripe = require('stripe')(process.env.STRIPE_SECRET_KEY);
const app = express();

// Important: Stripe requires the raw body buffer for signature verification
app.post('/api/webhooks/stripe', express.raw({ type: 'application/json' }), (req, res) => {
  const sig = req.headers['stripe-signature'];
  const endpointSecret = process.env.STRIPE_WEBHOOK_SECRET;

  let event;
  try {
    event = stripe.webhooks.constructEvent(req.body, sig, endpointSecret);
  } catch (err) {
    console.error(`❌ Webhook signature verification failed:`, err.message);
    return res.status(400).send(`Webhook Error: ${err.message}`);
  }

  // Handle the event
  switch (event.type) {
    case 'payment_intent.succeeded':
      const paymentIntent = event.data.object;
      console.log(`✅ PaymentIntent succeeded for ${paymentIntent.amount}!`);
      break;
    default:
      console.log(`Unhandled event type ${event.type}`);
  }

  res.json({ received: true });
});

app.listen(3000, () => console.log('Server listening on http://localhost:3000'));
```

### Step 2: Open the Tunnel

In a separate terminal, expose port 3000:

```bash
expose tunnel 3000
```

Output:
```text
🚀 Tunnel[localtunnel] started for localhost:3000
✓ Public URL: https://quick-mammals-sing.loca.lt
✓ Forwarding to: http://localhost:3000 (Ready)
✓ Provider: localtunnel
Press Ctrl+C to stop
```

### Step 3: Register Webhook in Stripe Dashboard

1. Navigate to the **Stripe Dashboard** $\rightarrow$ **Developers** $\rightarrow$ **Webhooks**.
2. Click **Add endpoint**.
3. In **Endpoint URL**, enter your Expose URL:
   ```text
   https://quick-mammals-sing.loca.lt/api/webhooks/stripe
   ```
4. Select the events you want to receive (e.g. `payment_intent.succeeded`).
5. Copy your **Signing secret** (`whsec_...`) into your local environment as `STRIPE_WEBHOOK_SECRET`.
6. Click **Send test event**.

You will instantly see the 200 OK response logged in your local console with signature verification passing!

---

## Avoiding Changing Webhook URLs with Self-Hosted Expose

On free public relays, tunnels receive an ephemeral random subdomain (e.g., `swift-falcon-4821.loca.lt`). Each time you restart the process, you have to update the webhook URL in the provider dashboard.

With Expose's built-in self-hosted relay, you can **reserve a permanent, persistent subdomain** without paying for commercial subscription tiers:

```bash
expose tunnel 3000 --server tunnel.mysite.com:7890 --subdomain stripe-webhooks
```

This guarantees that your public URL is always:
```text
https://stripe-webhooks.tunnel.mysite.com/api/webhooks/stripe
```

Now you can configure your Stripe or GitHub dashboard **once**, and it remains valid permanently across system reboots, coffee shop Wi-Fi changes, and development sessions.

---

## Quickstart: Testing GitHub Webhooks

1. Start your local server listening on port 8080.
2. Run Expose:
   ```bash
   expose tunnel 8080
   ```
3. Copy the generated HTTPS URL.
4. Go to your GitHub repository $\rightarrow$ **Settings** $\rightarrow$ **Webhooks** $\rightarrow$ **Add webhook**.
5. Set **Payload URL** to:
   ```text
   https://<your-subdomain>.loca.lt/github-webhook
   ```
6. Set **Content type** to `application/json`.
7. Enter a secret key and check **Just the push event**.
8. Click **Add webhook**. GitHub will send a test `ping` event. Expose forwards the request to your local listener with the `X-GitHub-Event: ping` and `X-Hub-Signature-256` headers intact.

