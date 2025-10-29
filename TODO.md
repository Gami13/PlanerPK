# TODO List for Production Deployment

## ✅ Already Planned
- [✅] Environment variables configuration
- [✅] Structured logging

## 🔴 HIGH PRIORITY (Critical for Production)

### Input Validation & Error Handling
- [ ] Validate `chosenColumn` input (check if within valid range, non-negative)
- [ ] Replace all `panic()` calls with proper error handling and HTTP error responses
- [ ] Add validation for all endpoint inputs
- [ ] Return user-friendly error messages (hide technical details from users)

### CORS Configuration
- [ ] Replace `Access-Control-Allow-Origin: *` with whitelist of allowed origins
- [ ] Load allowed origins from environment variables
- [ ] Set proper CORS headers (credentials, methods, headers)

### User Management
- [ ] Add DELETE `/api/account` endpoint for user deletion (GDPR compliance)
- [ ] Add ability for user to deactivate their account
- [ ] Implement soft delete vs hard delete strategy

### Missing API Endpoints
- [ ] `GET /api/health` - Health check endpoint for monitoring if the service is up 
- [ ] `GET /api/preview` - Preview what events will be created (before importing)


## 🟡 MEDIUM PRIORITY (Important for Quality)

### Rate Limiting
- [✅] Implement rate limiting per IP address
- [✅] Implement rate limiting per user (access token)
- [✅] Configure different limits for different endpoints

### Background Job Improvements
- [ ] Add retry logic for failed calendar updates
- [ ] Better error handling in the ticker goroutine


## 🟢 LOW PRIORITY (Nice to Have)

### Caching
- [ ] Cache timetable data in memory

### Documentation
- [ ] Create deployment guide
- [ ] Write user guide
- [ ] Document environment variables

### Deployment
- [ ] Create Dockerfile 
- [ ] Create docker-compose.yml 

