package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ameyzing09/rtr-user-auth-service/internal/config"
	"github.com/ameyzing09/rtr-user-auth-service/internal/domain/repositories"
	"github.com/ameyzing09/rtr-user-auth-service/internal/handlers"
	"github.com/ameyzing09/rtr-user-auth-service/internal/services"
	"github.com/ameyzing09/rtr-user-auth-service/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type IntegrationTestSuite struct {
	suite.Suite
	db         *gorm.DB
	router     *gin.Engine
	jwtService *utils.JWTService
}

func (suite *IntegrationTestSuite) SetupSuite() {
	// Use SQLite in-memory database for testing
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	suite.Require().NoError(err)
	
	suite.db = db
	
	// Create database wrapper
	database := &config.Database{DB: db}
	err = database.Migrate()
	suite.Require().NoError(err)
	
	// Initialize repositories
	tenantRepo := repositories.NewTenantRepository(db)
	userRepo := repositories.NewUserRepository(db)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(db)
	
	// Initialize JWT service
	suite.jwtService = utils.NewJWTService("test-secret-key-32-characters-long", 
		24*60*60*1000000000, // 24 hours in nanoseconds
		168*60*60*1000000000) // 168 hours in nanoseconds
	
	// Initialize services
	authService := services.NewAuthService(userRepo, refreshTokenRepo, tenantRepo, suite.jwtService)
	tenantService := services.NewTenantService(tenantRepo)
	
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	tenantHandler := handlers.NewTenantHandler(tenantService)
	
	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Public routes
	public := router.Group("/api/v1")
	public.POST("/auth/login", authHandler.Login)
	public.GET("/tenants/by-domain", tenantHandler.GetTenantByDomain)
	
	suite.router = router
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	sqlDB, err := suite.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

func (suite *IntegrationTestSuite) SetupTest() {
	// Clean up database before each test
	suite.db.Exec("DELETE FROM refresh_tokens")
	suite.db.Exec("DELETE FROM users")
	suite.db.Exec("DELETE FROM tenants")
}

func (suite *IntegrationTestSuite) TestTenantByDomainEndpoint() {
	// Create a test tenant directly in database
	tenant := map[string]interface{}{
		"id":        "123e4567-e89b-12d3-a456-426614174000",
		"name":      "Test Corp",
		"domain":    "test.com",
		"is_active": true,
	}
	
	result := suite.db.Table("tenants").Create(tenant)
	suite.Require().NoError(result.Error)
	
	// Test getting tenant by domain
	req, _ := http.NewRequest("GET", "/api/v1/tenants/by-domain?domain=test.com", nil)
	resp := httptest.NewRecorder()
	
	suite.router.ServeHTTP(resp, req)
	
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Test Corp", response["name"])
	assert.Equal(suite.T(), "test.com", response["domain"])
}

func (suite *IntegrationTestSuite) TestLoginEndpoint() {
	// Create a test tenant and user directly in database
	tenantID := "123e4567-e89b-12d3-a456-426614174000"
	userID := "123e4567-e89b-12d3-a456-426614174001"
	
	tenant := map[string]interface{}{
		"id":        tenantID,
		"name":      "Test Corp",
		"domain":    "test.com", 
		"is_active": true,
	}
	
	result := suite.db.Table("tenants").Create(tenant)
	suite.Require().NoError(result.Error)
	
	// Hash password for test user
	hashedPassword, err := utils.HashPassword("password123")
	suite.Require().NoError(err)
	
	user := map[string]interface{}{
		"id":         userID,
		"tenant_id":  tenantID,
		"email":      "test@test.com",
		"password":   hashedPassword,
		"first_name": "Test",
		"last_name":  "User",
		"role":       "CANDIDATE",
		"is_active":  true,
	}
	
	result = suite.db.Table("users").Create(user)
	suite.Require().NoError(result.Error)
	
	// Test login
	loginReq := map[string]interface{}{
		"tenant_id": tenantID,
		"email":     "test@test.com",
		"password":  "password123",
	}
	
	jsonBody, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	
	suite.router.ServeHTTP(resp, req)
	
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), response["access_token"])
	assert.NotEmpty(suite.T(), response["refresh_token"])
	assert.Equal(suite.T(), "Bearer", response["token_type"])
	
	// Verify user data in response
	userData := response["user"].(map[string]interface{})
	assert.Equal(suite.T(), "test@test.com", userData["email"])
	assert.Equal(suite.T(), "Test", userData["first_name"])
	assert.Equal(suite.T(), "User", userData["last_name"])
	assert.Equal(suite.T(), "CANDIDATE", userData["role"])
}

func (suite *IntegrationTestSuite) TestLoginWithInvalidCredentials() {
	// Test login with invalid credentials
	loginReq := map[string]interface{}{
		"tenant_id": "123e4567-e89b-12d3-a456-426614174000",
		"email":     "invalid@test.com",
		"password":  "wrongpassword",
	}
	
	jsonBody, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	
	suite.router.ServeHTTP(resp, req)
	
	assert.Equal(suite.T(), http.StatusUnauthorized, resp.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "invalid_credentials", response["error"])
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}