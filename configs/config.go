package configs

import (
	"os"
)

func ConfigSetup() {
	os.Setenv("DB_USERNAME", "postgres")
	os.Setenv("DB_PASSWORD", "qwe")
	os.Setenv("DB_HOST", "localhost:5432")
	os.Setenv("DB_NAME", "totalTest")
	os.Setenv("DB_POOL_MAXCONN", "5")
	os.Setenv("DB_POOL_MAXCONN_LIFETIME", "300")
	os.Setenv("NATS_HOSTS", "0.0.0.0:4222")
	os.Setenv("NATS_CLUSTER_ID", "test-cluster")
	os.Setenv("NATS_CLIENT_ID", "someClientID")
	os.Setenv("NATS_SUBJECT", "DemoVideoTestChannel")
	os.Setenv("NATS_DURABLE_NAME", "Replica-1")
	os.Setenv("NATS_ACK_WAIT_SECONDS", "30")

	os.Setenv("CACHE_SIZE", "10")
	os.Setenv("APP_KEY", "APPKEY777")
}
