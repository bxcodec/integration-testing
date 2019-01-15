package redis_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bxcodec/integration-testing/models"
	redisHandler "github.com/bxcodec/integration-testing/redis"
	"github.com/stretchr/testify/suite"
)

type redisHandlerSuite struct {
	RedisSuite
}

func TestRedisSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip test for redis repository")
	}
	redisHostTest := os.Getenv("REDIS_TEST_URL")
	if redisHostTest == "" {
		redisHostTest = "localhost:6379"
	}
	redisHandlerSuiteTest := &redisHandlerSuite{
		RedisSuite{
			Host: redisHostTest,
		},
	}
	suite.Run(t, redisHandlerSuiteTest)
}

func getItemByKey(client *redis.Client, key string) ([]byte, error) {
	return client.Get(key).Bytes()
}
func seedItem(client *redis.Client, key string, value interface{}) error {
	jybt, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return client.Set(key, jybt, time.Second*30).Err()
}
func (r *redisHandlerSuite) TestSet() {
	testkeynews := "news"
	repo := redisHandler.NewHandler(r.Client, time.Second*30)
	news := models.Category{
		ID:   1,
		Name: "News",
		Slug: "news",
	}
	err := repo.Set(testkeynews, news)
	require.NoError(r.T(), err)

	jbyt, err := getItemByKey(r.Client, testkeynews)
	require.NoError(r.T(), err)
	require.NotNil(r.T(), jbyt)
	var insertedData models.Category
	err = json.Unmarshal(jbyt, &insertedData)
	require.NoError(r.T(), err)
	assert.Equal(r.T(), news.ID, insertedData.ID)
	assert.Equal(r.T(), news.Name, insertedData.Name)
	assert.Equal(r.T(), news.Slug, insertedData.Slug)
}
func (r *redisHandlerSuite) TestGet() {
	testkeybola := "bola"
	bola := models.Category{Name: "Bola", Slug: "bola", ID: 2}
	err := seedItem(r.Client, testkeybola, bola)
	require.NoError(r.T(), err)

	repo := redisHandler.NewHandler(r.Client, time.Second*300)
	jbyt, err := repo.Get(testkeybola)
	require.NoError(r.T(), err)
	var res models.Category
	err = json.Unmarshal(jbyt, &res)
	require.NoError(r.T(), err)

	assert.Equal(r.T(), bola.ID, res.ID)
	assert.Equal(r.T(), bola.Name, res.Name)
	assert.Equal(r.T(), bola.Slug, res.Slug)
}
