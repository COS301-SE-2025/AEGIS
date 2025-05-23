package auth_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "golang.org/x/crypto/bcrypt"
    "aegis-api/services/login/auth"
)

// MockUserRepo is a mock implementation of the UserRepository
type MockUserRepo struct {
    mock.Mock
}

func (m *MockUserRepo) GetUserByEmail(email string) (*auth.User, error) {
    args := m.Called(email)
    return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockUserRepo) CreateUser(user *auth.User) error {
    args := m.Called(user)
    return args.Error(0)
}

func (m *MockUserRepo) UpdateUser(user *auth.User) error {
    args := m.Called(user)
    return args.Error(0)
}

func (m *MockUserRepo) GetUserByToken(token string) (*auth.User, error) {
    args := m.Called(token)
    return args.Get(0).(*auth.User), args.Error(1)
}

// TestLogin tests the login logic.
func TestLogin(t *testing.T) {
    mockRepo := new(MockUserRepo)

    // Mock password hash (bcrypt hashed password)
    passwordHash := "$2a$10$Q2mGcNpsODX5gVxsSPlQ5urHXZ35cJk75GT.d6a5ERIRwHEBd4EBO" // bcrypt hash of "Fireal@chemist123"

    // Prepare mock user data with the correct field names
    mockUser := &auth.User{
        ID:               "1234",
        Email:            "roy@aegis.dev",
        PasswordHash:     passwordHash, // Correct field name for password hash
        VerificationToken: "valid-token", // Correct field name for token
    }

    // Mock GetUserByEmail to return mockUser
    mockRepo.On("GetUserByEmail", "roy@aegis.dev").Return(mockUser, nil)

    // Prepare login request
    req := auth.LoginRequest{
        Email:    "roy@aegis.dev",
        Password: "Fireal@chemist123",
    }
    reqBody, err := json.Marshal(req)
    if err != nil {
        t.Fatal(err)
    }

    // Create a new HTTP request with the login endpoint
    request := httptest.NewRequest("POST", "/login", bytes.NewReader(reqBody))
    recorder := httptest.NewRecorder()

    // Call LoginHandler function
    handler := http.HandlerFunc(auth.LoginHandler)

    // Call the handler with the request and recorder
    handler.ServeHTTP(recorder, request)

    // Check the status code
    if recorder.Code != http.StatusOK {
        t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
    }

    // Check the response body
    var resp auth.LoginResponse
    err = json.NewDecoder(recorder.Body).Decode(&resp)
    if err != nil {
        t.Fatal(err)
    }

    // Check if the response contains the correct values
    if resp.Email != req.Email {
        t.Errorf("Expected email %s, got %s", req.Email, resp.Email)
    }
    if resp.Token == "" {
        t.Error("Expected non-empty token, got empty")
    }
}
