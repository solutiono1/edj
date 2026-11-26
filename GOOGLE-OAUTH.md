# Getting the Google OAuth Client ID

We use Google Identity Services. We need only the **Client ID**, never the Client Secret.

## Steps
1. Open https://console.cloud.google.com/
2. Create a project (name it `pagor`) or select an existing one.
3. Go to **APIs and Services > OAuth consent screen**.
   - User type: **External**
   - App name: `pagor`
   - User support email: your email
   - Developer email: your email
   - Scopes: `openid`, `email`, `profile`
   - Add your Google account under **Test users**
   - You can keep it in Testing mode while we build. The test user can sign in.
4. Go to **APIs and Services > Credentials**.
   - Click **Create Credentials > OAuth client ID**
   - Application type: **Web application**
   - Name: `pagor web`
   - Authorized JavaScript origins:
     - `http://localhost:8080`
      - `http://127.0.0.1:8080`
      - Do not add a raw WSL IP origin. Google rejects it as an invalid origin.
     - `https://pagor.finance` (for later)
   - Authorized redirect URIs:
     - `http://localhost:8080/login`
     - `http://localhost:8080/signup`
5. Copy the **Client ID** (a long string ending in `.apps.googleusercontent.com`).
6. In WSL:
   ```bash
   cd /home/solution/pagor
   cp .env.example .env
   nano .env
   ```
   Add: `GOOGLE_CLIENT_ID=your_client_id_here`
7. Restart with `./start.sh`.

The server verifies the ID token with Google's `tokeninfo` endpoint and checks the audience against this client ID. No client secret is needed.

## Current configuration
- Client ID: `146518974987-4oe4a18mkqirdu4fhh1ui0fuqkvu6gsg.apps.googleusercontent.com`
- Stored in `.env` as `GOOGLE_CLIENT_ID`.
- Authorized JavaScript origins must include:
  - `http://localhost:8080`
  - `https://pagor.finance` (when the public domain is live)
- Authorized redirect URIs are not required for the popup flow, but adding `/login` and `/signup` on each origin is harmless.

If the Google button renders but sign-in fails with an origin error, use the exact origin shown in the error. Google rejects non-loopback IP origins such as `http://172.18.204.175:8080` with **Invalid Origin: must end with a public top-level domain**. Do not add the WSL IP. Use `http://localhost:8080` and `http://127.0.0.1:8080` for local testing, and `https://pagor.finance` in production. The app renders the official button through Google Identity Services and stores only the returned session token in localStorage. No client secret and no cookie are used.

Do not commit `.env`. It is already in `.gitignore`.
