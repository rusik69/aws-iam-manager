# OAuth2 Proxy Setup with Google

This document explains how to configure OAuth2 Proxy with Google for the AWS IAM Manager application.

## Prerequisites

1. A Google account
2. Access to Google Cloud Console

## Google OAuth2 Setup

### 1. Create OAuth2 Credentials in Google Cloud Console

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Navigate to **APIs & Services** > **Credentials**
4. Click **Create Credentials** > **OAuth 2.0 Client IDs**
5. Configure the consent screen if not already done:
   - Choose **External** for user type
   - Fill in required fields (app name, user support email, etc.)
   - Add your domain to authorized domains
6. Create OAuth 2.0 Client ID:
   - Application type: **Web application**
   - Name: `AWS IAM Manager`
   - Authorized redirect URIs:
     - Add the URL for each deployment host: `http://your-server-ip/oauth2/callback`
     - For example: `http://172.19.112.136/oauth2/callback`
     - `http://localhost/oauth2/callback` (for local testing)

### 2. Configure Environment Variables

Copy `.env.example` to `.env` and update the OAuth2 settings:

```bash
cp .env.example .env
```

Update the following variables in `.env`:

```bash
# Generate a random 32-character cookie secret
OAUTH2_COOKIE_SECRET=$(openssl rand -base64 32)

# From Google Cloud Console OAuth2 credentials
OAUTH2_CLIENT_ID=your_client_id.apps.googleusercontent.com
OAUTH2_CLIENT_SECRET=your_client_secret_from_google

# Redirect URL (automatically updated during deployment)
OAUTH2_REDIRECT_URL=http://localhost/oauth2/callback
```

### 3. Configure Authorized Users

Edit `oauth2-proxy/authenticated-emails.txt` to add authorized email addresses:

```
user1@gmail.com
user2@company.com
admin@organization.org
```

Each email address should be on a separate line.

## Deployment

Deploy the application with OAuth2 Proxy using Kubernetes:

```bash
make deploy HOST=your-server USER=your-username
```

**Note**: The deployment process automatically updates the `OAUTH2_REDIRECT_URL` in the environment configuration on the target server to use the correct host (e.g., `http://your-server/oauth2/callback`). Make sure this URL is configured in your Google Cloud Console OAuth2 settings.

## Security Features

- **Route Protection**: All application routes are protected except health checks (`/ping`, `/health`, `/ready`)
- **Email-based Authorization**: Only users listed in `authenticated-emails.txt` can access the application
- **Google OAuth2**: Uses Google's secure OAuth2 flow for authentication
- **Cookie Security**: Secure cookie handling with HttpOnly flags
- **User Headers**: Passes authenticated user information to the backend application

## Testing

1. Navigate to your application URL
2. You should be redirected to Google OAuth2 login
3. After successful authentication, you'll be redirected back to the application
4. Only authorized users (listed in authenticated-emails.txt) will gain access

## Troubleshooting

### Common Issues

1. **Invalid redirect URI**: Make sure the redirect URL in Google Console matches `OAUTH2_REDIRECT_URL`
2. **Access denied**: Check if the user's email is listed in `oauth2-proxy/authenticated-emails.txt`
3. **Cookie errors**: Verify `OAUTH2_COOKIE_SECRET` is a 32-character string

### Logs

Check OAuth2 Proxy logs:
```bash
ssh user@server 'kubectl logs -n aws-iam-manager deployment/oauth2-proxy'
```

Check application logs:
```bash
ssh user@server 'kubectl logs -n aws-iam-manager deployment/aws-iam-manager'
```

## Architecture

```
Internet → Ingress → OAuth2 Proxy (Port 4180) → AWS IAM Manager (Port 8080)
```

- Kubernetes Ingress routes traffic to OAuth2 Proxy
- OAuth2 Proxy handles authentication and sits in front of the application
- All requests are authenticated before reaching the AWS IAM Manager
- User information is passed to the backend via headers