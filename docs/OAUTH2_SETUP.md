# OAuth2 Proxy Setup with Okta SSO

This guide walks you through setting up OAuth2 Proxy with Okta SSO for the AWS IAM Manager application.

## Overview

The AWS IAM Manager now supports authentication via Okta Single Sign-On (SSO) using OAuth2 Proxy. OAuth2 Proxy acts as a reverse proxy that handles authentication before requests reach the application.

## Architecture

```
User Browser → OAuth2 Proxy (port 4180) → AWS IAM Manager (port 8080)
                      ↓
                 Okta OIDC Provider
```

## Prerequisites

1. **Okta Account**: You need an Okta organization with admin access
2. **Domain**: A domain name where you'll host the application (for production)
3. **Docker & Docker Compose**: For running the application stack

## Step 1: Configure Okta Application

### 1.1 Create a New App Integration

1. Log into your Okta Admin Console
2. Navigate to **Applications** → **Applications**
3. Click **Create App Integration**
4. Select:
   - **Sign-in method**: OIDC - OpenID Connect
   - **Application type**: Web Application
5. Click **Next**

### 1.2 Configure Application Settings

**General Settings:**
- **App integration name**: `AWS IAM Manager`
- **App logo**: (optional)

**Grant Types:**
- ✅ Authorization Code
- ✅ Refresh Token

**Sign-in redirect URIs:**
- For local development: `http://localhost:4180/oauth2/callback`
- For production: `https://your-domain.com/oauth2/callback`

**Sign-out redirect URIs:**
- For local development: `http://localhost:4180/oauth2/sign_out`
- For production: `https://your-domain.com/oauth2/sign_out`

**Controlled access:**
- Choose appropriate access level (e.g., specific groups or users)

### 1.3 Save Client Credentials

After creating the application, save these values:
- **Client ID**: Found in the application's General tab
- **Client Secret**: Found in the application's General tab
- **Okta Domain**: Your Okta domain (e.g., `your-company.okta.com`)

## Step 2: Configure Environment Variables

### 2.1 Copy Environment Template

```bash
cp .env.example .env
```

### 2.2 Update Environment Variables

Edit the `.env` file with your specific values:

```bash
# AWS Configuration (existing)
AWS_ACCESS_KEY_ID=your_aws_access_key_here
AWS_SECRET_ACCESS_KEY=your_aws_secret_key_here
AWS_REGION=us-east-1

# Application Configuration
PORT=8080

# OAuth2 Proxy Configuration
# Generate with: openssl rand -base64 32
OAUTH2_PROXY_COOKIE_SECRET=your_32_character_random_string_here

# Okta OIDC Configuration
OKTA_ISSUER_URL=https://your-company.okta.com/oauth2/default
OKTA_CLIENT_ID=your_okta_client_id_here
OKTA_CLIENT_SECRET=your_okta_client_secret_here

# OAuth2 Proxy Redirect URL
OAUTH2_PROXY_REDIRECT_URL=http://localhost:4180/oauth2/callback

# Optional: Access Control
# OAUTH2_PROXY_ALLOWED_GROUPS=aws-iam-manager-users
# OAUTH2_PROXY_ALLOWED_EMAILS=admin@company.com,user@company.com
```

### 2.3 Generate Cookie Secret

Generate a secure cookie secret:

```bash
openssl rand -base64 32
```

## Step 3: Configure OAuth2 Proxy (Optional)

The `oauth2-proxy.cfg` file contains the OAuth2 Proxy configuration. You can modify it to:

- Restrict access to specific groups
- Change session timeouts
- Modify cookie settings
- Add custom headers

Example group restriction:
```
allowed_groups = ["aws-admin-group", "aws-users-group"]
```

## Step 4: Configure Okta Groups (Optional)

### 4.1 Create Groups in Okta

1. Navigate to **Directory** → **Groups**
2. Create groups like:
   - `aws-iam-manager-admins`
   - `aws-iam-manager-users`

### 4.2 Assign Users to Groups

1. Select a group
2. Click **Assign** → **Assign to People**
3. Add appropriate users

### 4.3 Configure Group Claims

1. Go to **Security** → **API** → **Authorization Servers**
2. Select **default** or create a custom authorization server
3. Go to **Claims** tab
4. Add a custom claim:
   - **Name**: `groups`
   - **Include in token type**: ID Token, Always
   - **Value type**: Groups
   - **Filter**: Regex: `.*` (or specific group pattern)

## Step 5: Deploy the Application

### 5.1 Local Development

```bash
# Start the application stack
docker-compose up --build

# The application will be available at:
# http://localhost:4180 (OAuth2 Proxy - main entry point)
# http://localhost:8080 (Direct app access - bypasses auth)
```

### 5.2 Production Deployment

For production, update:

1. **Domain configuration** in Okta app settings
2. **Redirect URLs** to use HTTPS
3. **Environment variables** for production values
4. **SSL/TLS termination** (use a reverse proxy like nginx)

Example production `docker-compose.yml` addition:

```yaml
services:
  oauth2-proxy:
    # ... existing configuration
    environment:
      # ... existing environment
      - OAUTH2_PROXY_REDIRECT_URL=https://your-domain.com/oauth2/callback
    # Add nginx or traefik for SSL termination
```

## Step 6: Access Control Configuration

### 6.1 Group-based Access Control

To restrict access to specific Okta groups, uncomment in `.env`:

```bash
OAUTH2_PROXY_ALLOWED_GROUPS=aws-iam-manager-admins,aws-iam-manager-users
```

### 6.2 Email-based Access Control

To restrict access to specific email addresses:

```bash
OAUTH2_PROXY_ALLOWED_EMAILS=admin@company.com,user@company.com
```

### 6.3 Application-level Access Control

You can also implement access control within the application using the middleware:

```go
// Example: Restrict specific endpoints to admin group
adminRoutes := api.Group("/admin")
adminRoutes.Use(middleware.RequireGroup("aws-iam-manager-admins"))
{
    adminRoutes.DELETE("/accounts/:accountId/users/:username", handler.DeleteUser)
}
```

## Step 7: Testing the Setup

### 7.1 Test Authentication Flow

1. Navigate to `http://localhost:4180`
2. You should be redirected to Okta login
3. After successful login, you should be redirected back to the application
4. Verify user information at: `http://localhost:4180/api/auth/user`

### 7.2 Test API Access

```bash
# This should require authentication
curl http://localhost:4180/api/accounts

# This should work without authentication (health check)
curl http://localhost:4180/health
```

## Troubleshooting

### Common Issues

**1. "Invalid client_id or client_secret"**
- Verify your Okta application credentials
- Ensure the client secret is correct
- Check that the application is active in Okta

**2. "Redirect URI mismatch"**
- Verify the redirect URI in Okta matches your environment variable
- Ensure the protocol (http/https) matches

**3. "Groups not appearing in token"**
- Verify groups claim is configured in Okta
- Check that users are assigned to the correct groups
- Ensure the groups claim is included in the ID token

**4. "Authentication loop"**
- Check cookie secret is properly set
- Verify cookie domain settings
- Check browser console for errors

### Debug Mode

Enable debug logging by adding to `oauth2-proxy.cfg`:

```
logging_level = "debug"
```

### Health Check Endpoints

- `/ping` - OAuth2 Proxy health check
- `/health` - Application health check
- `/oauth2/auth` - OAuth2 Proxy auth status

## Security Considerations

1. **Cookie Security**:
   - Use secure cookies in production (HTTPS)
   - Set appropriate cookie expiration times

2. **Session Management**:
   - Configure appropriate session timeouts
   - Consider using Redis for session storage in production

3. **Network Security**:
   - Don't expose the application port (8080) directly
   - Only expose OAuth2 Proxy port (4180)
   - Use HTTPS in production

4. **Access Control**:
   - Implement least-privilege access
   - Regularly review group memberships
   - Use application-level authorization for sensitive operations

## Production Considerations

1. **High Availability**:
   - Run multiple OAuth2 Proxy instances behind a load balancer
   - Use external session storage (Redis)

2. **Monitoring**:
   - Monitor authentication failures
   - Set up alerts for security events
   - Log user access for audit purposes

3. **Backup**:
   - Backup OAuth2 Proxy configuration
   - Document group and user assignments

## Additional Resources

- [OAuth2 Proxy Documentation](https://oauth2-proxy.github.io/oauth2-proxy/)
- [Okta OIDC Documentation](https://developer.okta.com/docs/guides/implement-grant-type/authcode/main/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)