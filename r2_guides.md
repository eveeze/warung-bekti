# Cloudflare R2 Setup Guide

This guide explains how to set up Cloudflare R2 for your Warung Backend storage and how to connect your custom domain `warungmanto.store`.

## 1. Create a Bucket ("warung-assets")

1.  Log in to the [Cloudflare Dashboard](https://dash.cloudflare.com/).
2.  Navigate to **R2** from the sidebar.
3.  Click **Create Bucket**.
4.  Name your bucket `warung-assets` (or similar) and create it.

## 2. Generate API Tokens

1.  In the R2 dashboard, look for **Manage R2 API Tokens** in the sidebar (right side).
2.  Click **Create API Token**.
3.  **Permissions**: Select **Admin Read & Write**.
4.  **TTL**: Select **Forever** (or valid for as long as you need).
5.  Click **Create API Token**.

**IMPORTANT**: Copy the following values immediately and save them safely:

- **Account ID**: Found on the R2 Overview page (top right).
- **Access Key ID**
- **Secret Access Key**

## 3. Connect Custom Domain (`warungmanto.store`)

Since your domain is on **Spaceship.com**, you first need to connect it to Cloudflare to use R2 Custom Domains.

### Step A: Add Domain to Cloudflare

1.  Go to the Cloudflare Dashboard Home.
2.  Click **Add a Site**.
3.  Enter `warungmanto.store`.
4.  Select the **Free Plan** and continue.
5.  Cloudflare will scan your DNS records. Click **Continue**.
6.  Cloudflare will show you two **Nameservers** (e.g., `bob.ns.cloudflare.com` and `alice.ns.cloudflare.com`). **Copy these.**

### Step B: Update Nameservers on Spaceship.com

1.  Log in to your **Spaceship.com** account.
2.  Go to **Domain List** and find `warungmanto.store`.
3.  Click **Manage** (or similar settings icon).
4.  Find the **DNS / Nameservers** section.
5.  Change from "Spaceship DNS" to **Custom DNS**.
6.  Paste the two Nameservers provided by Cloudflare.
7.  Save changes.
    - _Note: DNS propagation usually takes 10-20 minutes, but can take up to 24 hours._

### Step C: Connect Domain to R2 Bucket

Since your domain is now connected to Cloudflare:

1.  Go back to **R2** -> Select your bucket (`warung-assets`).
2.  Click the **Settings** tab.
3.  Find the **Public Access** section.
4.  Click **Custom Domains** -> **Connect Domain**.
5.  Enter `assets.warungmanto.store`.
6.  Click **Continue** -> **Connect Domain**.
7.  Wait for the status to become **Active** (green).

### Step D: Update .env

Once active, update your `.env` file to use your custom domain.

## 4. Configure Backend

Open your `.env` file in the project root and update the R2 section with your new credentials and domain:

```env
# Cloudflare R2 Configuration
R2_ACCOUNT_ID=your_account_id
R2_ACCESS_KEY_ID=your_access_key_id
R2_SECRET_ACCESS_KEY=your_secret_access_key
R2_BUCKET_NAME=warung-assets
R2_PUBLIC_URL=https://assets.warungmanto.store
```

## 5. Restart Backend

Restart your Go server to apply changes:

```bash
go run cmd/api/main.go
```

Your images will now be served from `https://assets.warungmanto.store/products/...`!
