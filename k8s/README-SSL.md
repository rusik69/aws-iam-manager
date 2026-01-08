# SSL Certificate Setup with Let's Encrypt

This setup automatically provisions SSL certificates using Let's Encrypt and cert-manager.

## Prerequisites

Before deploying with SSL support, ensure these components are installed in your Kubernetes cluster:

### 1. Nginx Ingress Controller

```bash
# Install nginx ingress controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.2/deploy/static/provider/cloud/deploy.yaml

# Wait for ingress controller to be ready
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s
```

### 2. cert-manager

```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.2/cert-manager.yaml

# Wait for cert-manager to be ready
kubectl wait --namespace cert-manager \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/instance=cert-manager \
  --timeout=90s
```

## Deployment

Once prerequisites are installed, deploy the application with your domain:

```bash
# Replace your-domain.com with your actual domain
make deploy HOST=your-domain.com

# Examples:
make deploy HOST=iammanager.example.com
make deploy HOST=aws-manager.mydomain.org
```

**Important**: The HOST variable will be used for:
- SSL certificate domain validation
- Ingress host routing

## SSL Certificate Process

1. **Certificate Request**: The Certificate resource requests a certificate from Let's Encrypt
2. **HTTP-01 Challenge**: cert-manager performs domain validation via HTTP-01 challenge
3. **Certificate Issuance**: Let's Encrypt issues the certificate
4. **Secret Creation**: Certificate is stored in `aws-iam-manager-tls` secret
5. **Ingress Configuration**: Nginx ingress uses the certificate for SSL termination

## Verification

Check certificate status:

```bash
# Check certificate status
kubectl get certificates -n aws-iam-manager

# Check certificate details
kubectl describe certificate aws-iam-manager-cert -n aws-iam-manager

# Check certificate secret
kubectl get secret aws-iam-manager-tls -n aws-iam-manager
```

## Troubleshooting

### Certificate Pending

If certificate status shows "Pending":

```bash
# Check cert-manager logs
kubectl logs -n cert-manager -l app=cert-manager

# Check certificate events
kubectl describe certificate aws-iam-manager-cert -n aws-iam-manager

# Check challenge status
kubectl get challenges -n aws-iam-manager
```

### Common Issues

1. **DNS not pointing to ingress**: Ensure your domain points to the ingress controller's external IP
2. **Firewall blocking port 80**: HTTP-01 challenge requires port 80 to be accessible
3. **Rate limiting**: Let's Encrypt has rate limits; use staging issuer for testing
4. **Invalid email**: Ensure the email in cert-manager.yaml is valid

### Using Staging Environment

For testing, use the staging issuer to avoid rate limits:

```bash
# Edit certificate.yaml to use staging issuer
# issuerRef:
#   name: letsencrypt-staging
#   kind: ClusterIssuer
```

## Security Notes

- Certificates auto-renew 15 days before expiry
- HTTP traffic is automatically redirected to HTTPS
- HTTPS is enforced for all routes
- Application authentication (admin username/password) applies over SSL