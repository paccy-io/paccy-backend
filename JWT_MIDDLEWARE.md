# JWT Authentication Middleware

## Overview

This document describes the JWT authentication middleware implementation for protecting user-specific routes in the Paccy backend API.

## Features

✅ **JWT Token Verification**: Validates Bearer tokens from Authorization headers  
✅ **User ID Authorization**: Ensures users can only access their own resources  
✅ **Combined Middleware**: Single middleware combining both JWT and user ID validation  
✅ **Comprehensive Logging**: Detailed logs for authentication events  
✅ **Error Handling**: Structured error responses for various failure scenarios

## Implementation

### Middleware Functions

#### 1. `JWTMiddleware(authService auth.AuthService)`

- Verifies JWT tokens from `Authorization: Bearer <token>` headers
- Validates token format and extracts claims
- Stores user claims and user ID in Fiber context
- Returns 401 for invalid/missing tokens

#### 2. `UserIDMiddleware()`

- Validates that route parameter `userID` matches authenticated user
- Prevents users from accessing other users' resources
- Returns 403 for authorization mismatches
- Requires JWT middleware to run first

#### 3. `CombinedJWTUserIDMiddleware(authService auth.AuthService)`

- Combines both JWT verification and user ID validation
- Single middleware for complete protection
- Recommended for user-specific endpoints

### Protected Routes

```go
// Protected user routes (require JWT + user ID validation)
protected := api.Group("/users")
protected.Use(middleware.CombinedJWTUserIDMiddleware(authService))
protected.Get("/:userID", userHandler.GetUser)
```

### Route Protection

- **Before**: `GET /users/:id` (unprotected)
- **After**: `GET /users/:userID` (protected with JWT + user authorization)

### Authentication Flow

1. **Client Request**: Include `Authorization: Bearer <jwt_token>` header
2. **JWT Verification**: Middleware validates token and extracts claims
3. **User ID Check**: Middleware compares route `userID` with token user ID
4. **Authorization**: Allow access only if IDs match
5. **Context**: Store user claims for handler use

### Error Responses

```json
// Missing Authorization header
{
  "success": false,
  "error": "Authorization header required",
  "code": 401
}

// Invalid token format
{
  "success": false,
  "error": "Invalid authorization header format",
  "code": 401
}

// Invalid/expired token
{
  "success": false,
  "error": "Invalid or expired token",
  "code": 401
}

// User ID mismatch
{
  "success": false,
  "error": "Access denied: you can only access your own resources",
  "code": 403
}
```

### Usage Example

```bash
# 1. Login to get JWT token
curl -X POST http://localhost:3000/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password"}'

# Response: {"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}

# 2. Access protected user endpoint
curl -X GET http://localhost:3000/users/YOUR_USER_ID \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# ✅ Success: Returns user data if userID matches token
# ❌ Error 403: If userID doesn't match authenticated user
# ❌ Error 401: If token is invalid/missing
```

### Handler Context Access

Handlers can access authenticated user information:

```go
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
    // Get user claims from JWT middleware
    userClaims := c.Locals("user_claims").(*jwt.Claims)
    authenticatedUserID := c.Locals("user_id").(string)

    // User is pre-authorized to access this resource
    // Route userID parameter matches authenticated user
    // ... handler logic
}
```

### Logging

All authentication events are logged with service prefix `[auth]`:

```bash
[auth] INFO: JWT token verified successfully user_id=64f1a2b3c4d5e6f7g8h9i0j1 request_id=abc123
[auth] WARN: User ID mismatch - unauthorized access attempt route_user_id=different_id authenticated_user_id=user_id request_id=abc123
[auth] ERROR: JWT token verification failed error="token is expired" request_id=abc123
```

## Security Features

- **Bearer Token Validation**: Strict JWT format checking
- **User Authorization**: Prevents cross-user data access
- **Request Tracking**: Unique request IDs for all auth events
- **Comprehensive Logging**: Full audit trail of authentication attempts
- **Error Standardization**: Consistent error response format

## Integration

The middleware is fully integrated with the existing:

- ✅ **Logging System**: Uses service-specific `[auth]` prefix
- ✅ **Discord Webhooks**: Authentication events can trigger notifications
- ✅ **Error Handling**: Consistent with application-wide error responses
- ✅ **Request Tracking**: Works with existing request ID middleware

## Future Enhancements

- Rate limiting for failed authentication attempts
- Role-based access control (RBAC) middleware
- API key authentication support
- OAuth2/OpenID Connect integration
