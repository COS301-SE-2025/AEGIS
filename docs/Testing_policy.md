# AEGIS Testing Policy

## Advanced Evidence Gathering and Investigation System
**Version:** 1.0  
**Last Updated:** September 2025  
**Team:** Incident Intel

---

## 1. Overview

This document outlines the comprehensive testing strategy for AEGIS, a secure containerized platform for digital forensics and incident response (DFIR). Given the critical nature of evidence handling and chain-of-custody requirements, our testing policy ensures system reliability, security, and compliance.

---

## 2. Testing Technologies

### 2.1 Backend Testing Stack

**Primary Framework:** Go Testing (`testing` package)
- **Justification:** Native Go testing provides excellent performance, is well-integrated with the Go ecosystem, and requires no external dependencies for basic testing.

**Key Libraries:**
- `github.com/stretchr/testify/require` - Assertion library for cleaner test assertions
- `github.com/stretchr/testify/assert` - Additional assertion helpers
- `github.com/stretchr/testify/mock` - Mocking framework for dependencies
- `net/http/httptest` - HTTP testing utilities for API endpoints
- `github.com/gin-gonic/gin` - Web framework with built-in testing support
- `gorm.io/gorm` - ORM with test database utilities

**Test Types Implemented:**
- Unit Tests
- Integration Tests
- Security Tests
- Performance Tests
- Reliability Tests


### 2.2 CI/CD Platform

**GitHub Actions**
- **Justification:** 
  - Native integration with GitHub repositories
  - Free for public repositories, cost-effective for private
  - Extensive marketplace of pre-built actions
  - Parallel job execution for faster builds
  - Easy configuration with YAML
  - Built-in secrets management
  - Docker container support

---

## 3. Testing Procedure

### 3.1 Development Workflow

```
Feature Branch → Development → Main
     ↓              ↓            ↓
  Local Tests   All Tests    All Tests
                + Lint       + Deploy
                + Coverage   
```

**Branch Strategy:**

1. **Feature Branches** (`feature/*`, `bugfix/*`)
   - Developers write tests alongside feature implementation
   - All unit tests must pass locally before PR
   - Minimum 70% code coverage for new code

2. **Development Branch** (`develop`)
   - Unit tests and integration tests run automatically on push
   - Linting and code quality checks enforced
   - Test coverage reports generated
   - Docker builds verified
   - Security scans (Trivy, Gosec)
   - Performance and reliability tests run on scheduled basis
   - Manual QA verification before merging to main

3. **Main Branch** (`main`)
   - Production-ready code only
   - All automated tests must pass
   - Deployment to on-site server
   - Post-deployment smoke tests


### 3.2 Code Review Requirements

Pull requests must meet the following criteria:

- ✅ All automated tests pass
- ✅ Code coverage ≥ 75% overall
- ✅ New features have corresponding tests
- ✅ Integration tests for API changes
- ✅ No critical security vulnerabilities
- ✅ Linting passes with zero errors
- ✅ At least one peer review approval

---

## 4. Test Categories & Requirements

### 4.1 Unit Tests

**Purpose:** Test individual functions and methods in isolation

**Requirements:**
- Must mock all external dependencies (database, IPFS, external APIs)
- Should execute in < 100ms per test
- Coverage target: ≥ 80% for business logic

**Example Structure:**
```go
// services/evidence/metadata/service_test.go
func TestUploadEvidence_Success(t *testing.T) {
    // Arrange
    mockRepo := new(MockRepository)
    mockIPFS := new(MockIPFSClient)
    svc := NewService(mockRepo, mockIPFS)
    
    // Act
    err := svc.UploadEvidence(validRequest)
    
    // Assert
    require.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

### 4.2 Integration Tests

**Purpose:** Test interaction between components (e.g., API → Service → Database)

**Requirements:**
- Use test database with migrations
- Clean database state between tests
- Test complete request/response cycles
- Test HTTP endpoints with various inputs
- Test authentication and authorization
- Test request validation and error responses
- Coverage target: ≥ 70% for integration paths

**Example Structure:**
```go
// integration_test/evidence_test.go
func TestEvidenceUploadFlow(t *testing.T) {
    // Setup test database and router
    testDB := setupTestDB(t)
    defer cleanupTestDB(testDB)
    
    router := setupRouter(testDB)
    server := httptest.NewServer(router)
    defer server.Close()
    
    // Test full upload flow including HTTP request
    w := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/api/v1/evidence", body)
    req.Header.Set("Authorization", "Bearer "+token)
    
    router.ServeHTTP(w, req)
    require.Equal(t, http.StatusCreated, w.Code)
}
```

### 4.3 Security Tests

**Purpose:** Verify security controls and prevent vulnerabilities

**Requirements:**
- Test SQL injection prevention
- Test XSS prevention
- Test CSRF protection
- Test authentication bypass attempts
- Test authorization boundaries
- Test input validation and sanitization
- Run automated security scans (Trivy, Gosec)



### 4.4 Reliability Tests

**Purpose:** Ensure system handles high load and concurrent operations

**Requirements:**
- Simulate concurrent users (10-100)
- Test endpoint performance under stress
- Error rate must be < 10% under normal load
- Test system recovery from failures

**Example Structure:**
```go
concurrentUsers := 10
	totalRequests := 50

	for _, ep := range endpoints {
		t.Run(ep.Name, func(t *testing.T) {
			result := runStressTest(t, ep, concurrentUsers, totalRequests, server)
			t.Logf("[%s] Requests: %d | Errors: %d | ErrorRate: %.2f%% | Duration: %s",
				result.Name, result.Total, result.Failure, result.ErrorRate, result.Duration.String())

			// Assert low error rate
			if result.ErrorRate > 5 {
				t.Errorf("High error rate: %.2f%% for endpoint %s", result.ErrorRate, result.Name)
			}
		})
	}
```




## 5. Test Coverage Requirements

### 5.1 Overall Coverage Targets

| Component | Minimum Coverage |
|-----------|-----------------|
| Business Logic | 75% |
| API Handlers | 75% |
| Database Repositories | 70% |
| React Components | 75% |
| Frontend Utilities | 70% |
| Utilities | 60% |
| **Overall Project** | **75%** | 


## 6. Responsibilities


### 6.1 Shared Responsibilities

All team members share responsibility for:
- Maintaining test quality and reliability
- Identifying and eliminating flaky tests
- Improving test coverage over time
- Participating in test strategy discussions
- Reviewing test-related pull requests
- Reporting test infrastructure issues

### 6.2 Escalation Path

**Test Failures:**
1. Developer fixes own test failures immediately
2. Integration Engineer notified if CI/CD issues
3. Team discussion for systemic problems

**Coverage Issues:**
1. Developer adds tests to meet requirement

