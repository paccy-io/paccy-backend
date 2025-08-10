# Discord Webhook Integration

This project includes Discord webhook integration for real-time monitoring and notifications.

## Features

### 🚨 **Automatic Error Alerts**
- All ERROR and FATAL level logs are automatically sent to Discord
- Rich embeds with error context, user IDs, and system information
- Color-coded alerts (red for errors, dark red for fatal)

### 🚀 **Startup Notifications**
- Automatic notification when the application starts/restarts
- Includes deployment information:
  - Application version
  - Environment (development/production)
  - Server port
  - Hostname
  - Startup timestamp

### 👤 **User Registration Notifications**
- Automatic notification when new users register successfully
- Includes user information:
  - Username
  - Email address
  - User ID
  - Registration timestamp

### 📧 **Email Verification Notifications**
- Automatic notification when users verify their email addresses
- Includes verification information:
  - Email address
  - User ID
  - Verification timestamp

### 📨 **Email Delivery Notifications**
- Automatic notification for every email sent (success or failure)
- **Success notifications** include:
  - Recipient email address
  - Email subject and template
  - Provider used (Resender/Gmail SMTP)
  - User ID and timestamp
- **Failure notifications** include:
  - All success fields plus error details
  - Helps monitor email delivery issues

## Setup Instructions

1. **Create Discord Webhook**
   - Go to your Discord server
   - Navigate to Server Settings → Integrations → Webhooks
   - Click "New Webhook"
   - Choose the channel where you want notifications
   - Copy the webhook URL

2. **Configure Environment Variable**
   ```bash
   DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN
   ```

3. **Enable in Application**
   - The webhook is automatically enabled when `DISCORD_WEBHOOK_URL` is set
   - No additional configuration required

## Notification Examples

### Startup Notification
```
🚀 Paccy Started Successfully

🟢 Paccy Deployment
Application has been deployed/restarted successfully

Version: 1.0.0
Environment: production
Port: 3000
Hostname: server-01
Startup Time: 2025-08-10 13:22:51 UTC
```

### User Registration Notification
```
👤 New User Registration

🆕 User Registered
A new user has successfully registered

Username: johndoe
Email: john@example.com
User ID: 6898738db444555c0ad490d4
Registration Time: 2025-08-10 13:25:17 UTC
```

### Email Verification Notification
```
✅ Email Verified

📧 Email Verification Complete
User has successfully verified their email address

Email: john@example.com
User ID: 6898738db444555c0ad490d4
Verification Time: 2025-08-10 13:30:45 UTC
```

### Email Success Notification
```
📧 Email Sent Successfully

✅ Email Delivery Success
Email has been delivered successfully

Recipient: john@example.com
Subject: Welcome to Paccy
Template: welcome_email
Provider: RESENDER
User ID: 6898738db444555c0ad490d4
Sent At: 2025-08-10 13:35:22 UTC
```

### Email Failure Notification
```
⚠️ Email Delivery Failed

❌ Email Delivery Failure
Failed to deliver email to recipient

Recipient: john@example.com
Subject: Welcome to Paccy
Template: welcome_email
Provider: RESENDER
User ID: 6898738db444555c0ad490d4
Error: SMTP connection timeout
Failed At: 2025-08-10 13:35:22 UTC
```

### Error Alert
```
🚨 AUTH Error Alert

[AUTH] ERROR
Failed login attempt detected

user_id: 456
ip: 192.168.1.1
error: invalid credentials
timestamp: 2025-08-10T13:25:30Z
```

## Log Levels

- **ERROR**: Sent to Discord automatically
- **FATAL**: Sent to Discord automatically  
- **INFO/DEBUG/WARN**: Only logged to console (not sent to Discord)

## Performance

- **Asynchronous**: Webhook calls are non-blocking
- **Fail-Safe**: Webhook failures don't affect application performance
- **Rate Limited**: Discord webhooks have built-in rate limiting

## Security

- Webhook URL should be kept secure (like any API key)
- Only ERROR and FATAL logs are sent (no sensitive DEBUG information)
- User data is included only when relevant to the error context

## Testing

Use the test file to verify webhook functionality:
```bash
go run examples/test_startup_notification.go
```

This will test all webhook types:
- Startup notification (green)
- User registration notification (blue) 
- Email verification notification (green)
- Email success notification (green)
- Email failure notification (orange/yellow)
- Regular logging output

Make sure to replace `YOUR_DISCORD_WEBHOOK_URL_HERE` with your actual webhook URL.

## Color Coding System

- 🟢 **Green (#5763719)**: Success events (startup, email verification, email delivery success)
- 🔵 **Blue (#3447003)**: User events (registration)  
- 🟡 **Orange/Yellow (#16776960)**: Warning events (email delivery failures)
- 🔴 **Red (#15158332)**: Error events (system issues, fatal errors)

## Monitoring Coverage

Your Discord webhook provides comprehensive monitoring of:

### 📊 **Application Health**
- Server deployments and restarts
- System errors and fatal issues
- Real-time application status

### 👥 **User Activity** 
- New user registrations with details
- Email verification completions
- User engagement tracking

### 📬 **Email System Monitoring**
- Every email sent (success/failure)
- Email provider performance (Resender/Gmail SMTP)
- Delivery issues and error debugging
- Template and subject tracking

This gives you complete visibility into your application's user lifecycle and email delivery system!
