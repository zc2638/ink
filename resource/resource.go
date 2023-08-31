// Copyright © 2023 zc2638 <zc2638@qq.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resource

import (
	"embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/httpfs"

	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
)

// EmbedDatabase contains the database directory
//
//go:embed database
var EmbedDatabase embed.FS

func MigrateDatabase(driver, dsn string) error {
	driver = strings.ToLower(driver)
	databaseURL := CleanDSNScheme(driver, dsn)
	if driver == "sqlite3" {
		driver = "sqlite"
	}

	source, err := httpfs.New(http.FS(EmbedDatabase), fmt.Sprintf("database/migrations/%s", driver))
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("httpfs", source, databaseURL)
	if err != nil {
		return err
	}
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func CleanDSNScheme(scheme, dsn string) string {
	prefix := scheme + "://"
	return prefix + strings.TrimPrefix(dsn, prefix)
}
