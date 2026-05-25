package repositories_test

// Repository integration tests require a live PostgreSQL database.
// These tests are intentionally skipped until a test DB is provisioned.
//
// TODO: use testcontainers-go to spin up a Postgres container and run goose
//       migrations before each test suite.  Each test should run in its own
//       transaction rolled back at the end to keep tests isolated.
//
// Example setup (to be implemented):
//
//   func TestMain(m *testing.M) {
//       ctx := context.Background()
//       container, connStr := startPostgres(ctx)
//       defer container.Terminate(ctx)
//       runMigrations(connStr)
//       pool = connectPool(ctx, connStr)
//       os.Exit(m.Run())
//   }
