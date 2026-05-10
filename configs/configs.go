package configs

import (
	"sync"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

// Config mirrors every env-var exactly; all public. so if any error can check using os.GetEnv()
type config struct {
	// App
	APP_NAME   string
	APP_PORT   int
	LOG_LEVEL  string
	MACHINE_ID int
	ENV        string

	// Authentication
	JWT_SECRET               string
	INTERNAL_PASSPORT_SECRET string
	GATEWAY_SECRET           string

	// PostgreSQL
	POSTGRES_HOST     string
	POSTGRES_PORT     string
	POSTGRES_USER     string
	POSTGRES_PASSWORD string
	POSTGRES_DB       string
}

var Config *config
var once sync.Once

func InitializeConfigs() {
	once.Do(func() {

		err := godotenv.Load()
		if err != nil {
			err = godotenv.Load("../.env") // up one level
		}

		if err != nil {
			logrus.Error("Unable to initialize configs. No .env file found!")
		}

		Config = &config{}
		if err := envconfig.Process("", Config); err != nil {
			panic("config: " + err.Error())
		}
	})
}
