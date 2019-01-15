package mysql_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/bxcodec/integration-testing/models"
	repoHandler "github.com/bxcodec/integration-testing/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type mysqlCategorySuiteTest struct {
	MysqlSuite
}

func TestCategorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip category mysql repository test")
	}
	dsn := os.Getenv("MYSQL_TEST_URL")
	if dsn == "" {
		dsn = "root:root-pass@tcp(localhost:33060)/testing?parseTime=1&loc=Asia%2FJakarta&charset=utf8mb4&collation=utf8mb4_unicode_ci"
	}
	categorySuite := &mysqlCategorySuiteTest{
		MysqlSuite{
			DSN:                     dsn,
			MigrationLocationFolder: "migrations",
		},
	}

	suite.Run(t, categorySuite)
}

func (s *mysqlCategorySuiteTest) SetupTest() {
	log.Println("Starting a Test. Migrating the Database")
	err, _ := s.Migration.Up()
	require.NoError(s.T(), err)
	log.Println("Database Migrated Successfully")
}

func (s *mysqlCategorySuiteTest) TearDownTest()  {
	log.Println("Finishing Test. Dropping The Database")
	err, _ := s.Migration.Down()
	require.NoError(s.T(), err)
	log.Println("Database Dropped Successfully")
}

// https://blevesearch.com/news/Deferred-Cleanup,-Checking-Errors,-and-Potential-Problems/
func Close(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Fatal(err)
	}
}
func getCategoryByID(t *testing.T, DBconn *sql.DB, id int64) *models.Category {
	res := &models.Category{}

	query := `SELECT id, name, slug, created_at, updated_at FROM category WHERE id=?`

	row := DBconn.QueryRow(query, id)
	err := row.Scan(
		&res.ID,
		&res.Name,
		&res.Slug,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	if err == nil {
		return res
	}
	if err != sql.ErrNoRows {
		require.NoError(t, err)
	}
	return nil
}

func getMockArrCategory() []models.Category {
	return []models.Category{
		models.Category{
			ID:   1,
			Name: "Tekno",
			Slug: "tekno",
		},
		models.Category{
			ID:   2,
			Name: "Bola",
			Slug: "bola",
		},
		models.Category{
			ID:   3,
			Name: "Asmara",
			Slug: "asmara",
		},
		models.Category{
			ID:   4,
			Name: "Celebs",
			Slug: "celebs",
		},
	}
}

func seedCategoryData(t *testing.T, DBConn *sql.DB) {
	arrCategories := getMockArrCategory()
	query := `INSERT category SET id=?, name=?, slug=?, created_at=?, updated_at=?`
	stmt, err := DBConn.Prepare(query)
	require.NoError(t, err)
	defer Close(stmt)
	for _, category := range arrCategories {
		_, err := stmt.Exec(category.ID, category.Name, category.Slug, time.Now(), time.Now())
		require.NoError(t, err)
	}
}

func (m *mysqlCategorySuiteTest) TestStore() {

	// Prepare Steps
	repo := repoHandler.NewHandler(m.DBConn)

	type testCase struct {
		Name           string
		Payload        *models.Category
		ExpectedResult error
	}

	arrTestcase := []testCase{
		testCase{
			Name: "store-success",
			Payload: &models.Category{
				Name: "News",
				Slug: "news",
			},
			ExpectedResult: nil,
		},
		testCase{
			Name: "store-conflict",
			Payload: &models.Category{
				Name: "News",
				Slug: "news",
			},
			ExpectedResult: errors.New("Category is Duplicated"),
		},
	}

	for _, tc := range arrTestcase {
		m.T().Run(tc.Name, func(t *testing.T) {
			err := repo.Store(context.Background(), tc.Payload)
			require.Equal(m.T(), tc.ExpectedResult, err)
			if err == nil {
				assert.NotZero(m.T(), tc.Payload.ID)
				res := getCategoryByID(m.T(), m.DBConn, tc.Payload.ID)
				assert.NotNil(m.T(), res)
				assert.Equal(m.T(), tc.Payload.Slug, res.Slug)
			}
		})
	}
}

func (m *mysqlCategorySuiteTest) TestFetch() {
	repo := repoHandler.NewHandler(m.DBConn)
	seedCategoryData(m.T(), m.DBConn)

	type testCase struct {
		Name           string
		Filter         models.Filter
		ExpectedResult []models.Category
	}

	arrTestcase := []testCase{
		testCase{
			Name: "fetch-without-cursor-keyword-success",
			Filter: models.Filter{
				Num: 3,
			},
			ExpectedResult: []models.Category{
				models.Category{
					ID:   4,
					Name: "Celebs",
					Slug: "celebs",
				},
				models.Category{
					ID:   3,
					Name: "Asmara",
					Slug: "asmara",
				},
				models.Category{
					ID:   2,
					Name: "Bola",
					Slug: "bola",
				},
			},
		},
		testCase{
			Name: "fetch-with-cursor",
			Filter: models.Filter{
				Num:    3,
				Cursor: "3",
			},
			ExpectedResult: []models.Category{
				models.Category{
					ID:   2,
					Name: "Bola",
					Slug: "bola",
				},
				models.Category{
					ID:   1,
					Name: "Tekno",
					Slug: "tekno",
				},
			},
		},
		testCase{
			Name: "fetch-with-keyword",
			Filter: models.Filter{
				Num:     3,
				Keyword: "asm",
			},
			ExpectedResult: []models.Category{
				models.Category{
					ID:   3,
					Name: "Asmara",
					Slug: "asmara",
				},
			},
		},
	}

	for _, tc := range arrTestcase {
		m.T().Run(tc.Name, func(t *testing.T) {
			res, err := repo.Fetch(context.Background(), tc.Filter)
			require.NoError(t, err)
			require.Equal(t, len(tc.ExpectedResult), len(res), tc.Name)
			for i, item := range res {
				assert.Equal(t, tc.ExpectedResult[i].ID, item.ID)
				assert.Equal(t, tc.ExpectedResult[i].Name, item.Name)
				assert.Equal(t, tc.ExpectedResult[i].Slug, item.Slug)
			}
		})
	}
}
func (m *mysqlCategorySuiteTest) TestGetByID() {
	// Prepare
	mockCategory := getMockArrCategory()[0]
	seedCategoryData(m.T(), m.DBConn)
	repo := repoHandler.NewHandler(m.DBConn)

	// Test the function
	res, err := repo.GetByID(context.Background(), mockCategory.ID)

	// Evaluate the results
	require.NoError(m.T(), err)
	assert.Equal(m.T(), mockCategory.ID, res.ID)
	assert.Equal(m.T(), mockCategory.Name, res.Name)
	assert.Equal(m.T(), mockCategory.Slug, res.Slug)
}

func (m *mysqlCategorySuiteTest) TestGetBySlug() {
	// Prepare
	mockCategory := getMockArrCategory()[0]
	seedCategoryData(m.T(), m.DBConn)
	repo := repoHandler.NewHandler(m.DBConn)

	// Test the function
	res, err := repo.GetBySlug(context.Background(), mockCategory.Slug)

	// Evaluate the results
	require.NoError(m.T(), err)
	assert.Equal(m.T(), mockCategory.ID, res.ID)
	assert.Equal(m.T(), mockCategory.Name, res.Name)
	assert.Equal(m.T(), mockCategory.Slug, res.Slug)
}

func (m *mysqlCategorySuiteTest) TestUpdate() {
	// Prepare
	mockCategory := getMockArrCategory()[0]
	seedCategoryData(m.T(), m.DBConn)
	repo := repoHandler.NewHandler(m.DBConn)

	mockCategory.Name = "Teknologi" // previously only "tekno"

	// Test the function
	err := repo.Update(context.Background(), &mockCategory)

	// Evaluate the results
	require.NoError(m.T(), err)
	res := getCategoryByID(m.T(), m.DBConn, mockCategory.ID)
	assert.NotNil(m.T(), res)
	assert.Equal(m.T(), mockCategory.ID, res.ID)
	assert.Equal(m.T(), mockCategory.Name, res.Name)
	assert.Equal(m.T(), mockCategory.Slug, res.Slug)
}

func (m *mysqlCategorySuiteTest) TestDelete() {
	// Prepare
	mockCategory := getMockArrCategory()[0]
	seedCategoryData(m.T(), m.DBConn)
	repo := repoHandler.NewHandler(m.DBConn)

	// Test the function
	err := repo.Delete(context.Background(), fmt.Sprintf("%d", mockCategory.ID))

	// Evaluate the results
	require.NoError(m.T(), err)
	res := getCategoryByID(m.T(), m.DBConn, mockCategory.ID)
	assert.Nil(m.T(), res) // because already deleted the category should be nil
}
