package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server 		ServerConfig
	Database 	DatabaseConfig
	JWT			JWTConfig
	App 		AppConfig
	Hashing 	HashingConfig
//  Redis 		RedisConfig

}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	URI		 string
	Host	 string
	Port	 string
	Username string
	Password string
	DBName   string
}

type JWTConfig struct {
	SecretKey string
	ExpiresIn int
}

type AppConfig struct {
	Name         	string
	Version      	string
	Environment  	string
	LoggingLevel 	string
}

// argon2id parameters
type HashingConfig struct {
	Algorithm 		string //"argon2id"
	Memory          uint32
	Iterations      uint32
	Parallelism     uint8
	SaltLength      uint32
	KeyLength       uint32
}

func LoadConfig() *Config { // LoadConfig function initializes and returns a pointer!! to Config struct
	return &Config{ // returns the pointer of itself
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "3000"),
			Host: getEnv("SERVER_HOST", "localhost"),
		},
		Database: DatabaseConfig{
			URI:      getEnv("DATABASE_URI", "mongodb://localhost:27017"),
			Host:     getEnv("DATABASE_HOST", "localhost"),
			Port:     getEnv("DATABASE_PORT", "27017"),
			Username: getEnv("DATABASE_USERNAME", "admin"),
			Password: getEnv("DATABASE_PASSWORD", "admin"),
			DBName:   getEnv("DATABASE_NAME", "paccy"),
	},
		JWT: JWTConfig{
			SecretKey: getEnv("JWT_SECRET_KEY", "your_secret_key"),
			ExpiresIn: getEnvAsInt("JWT_EXPIRES_IN", 3600), // default 1 hour
		},
		App: AppConfig{
			Name:          getEnv("APP_NAME", "Paccy"),
			Version:       getEnv("APP_VERSION", "1.0.0"),
			Environment:   getEnv("APP_ENVIRONMENT", "development"),
			LoggingLevel:  getEnv("APP_LOGGING_LEVEL", "info"),
		},
		Hashing: HashingConfig{
			Algorithm:   getEnv("HASHING_ALGORITHM", "argon2id"),
			Memory:      uint32(getEnvAsInt("HASHING_MEMORY", 65536)),
			Iterations:  uint32(getEnvAsInt("HASHING_ITERATIONS", 3)),
			Parallelism: uint8(getEnvAsInt("HASHING_PARALLELISM", 1)),
			SaltLength:  uint32(getEnvAsInt("HASHING_SALT_LENGTH", 16)),
			KeyLength:   uint32(getEnvAsInt("HASHING_KEY_LENGTH", 32)),
		}}
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

func (cfg *Config) GetDatabaseURI() string {
	if cfg == nil || cfg.Database.URI == "" {
		return "mongodb://localhost:27017" // default URI if not set
	}
	return cfg.Database.URI 
}

func (cfg *Config) GetDatabaseName() string {
	if cfg == nil || cfg.Database.DBName == "" {
		return "paccy" // default database name if not set
	}
	return cfg.Database.DBName
}

func IsProduction(cfg *Config) bool {
	if cfg == nil || cfg.App.Environment == "" {
		return false // default to development if not set
	}
	return cfg.App.Environment == "production"
}



