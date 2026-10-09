//go:build go1.27

package stdlib_test

import (
	"fmt"
	"testing"
	"uuid"
)

// BenchmarkStringArrayScanResultDirect scans a PostgreSQL text[] result directly into a
// []string via the new driver.RowsColumnScanner interface (no SQLScanner wrapper). This
// is the new path enabled by Go 1.27.
func BenchmarkStringArrayScanResultDirect(b *testing.B) {
	db := openDB(b)
	defer closeDB(b, db)

	query := benchStringArraySelectSQL()
	b.ResetTimer()

	for b.Loop() {
		var result []string
		err := db.QueryRow(query).Scan(&result)
		if err != nil {
			b.Fatal(err)
		}
		if len(result) != benchStringArraySize {
			b.Fatalf("Expected %d, got %d", benchStringArraySize, len(result))
		}
	}
}

type BenchRowNative struct {
	ID       uuid.UUID
	ParentID *uuid.UUID
	Values   []int32
}

// BenchmarkSelectRowsScanNative scans into destinations that ScanColumn passes
// to the pgx type map instead of database/sql.
func BenchmarkSelectRowsScanNative(b *testing.B) {
	db := openDB(b)
	defer closeDB(b, db)

	rowCounts := getSelectRowsCounts(b)

	for _, rowCount := range rowCounts {
		b.Run(fmt.Sprintf("%d rows", rowCount), func(b *testing.B) {
			br := &BenchRowNative{}
			for b.Loop() {
				rows, err := db.Query("select md5(n::text)::uuid, md5((n + 1)::text)::uuid, array[n, n + 1]::int4[] from generate_series(1, $1) n", rowCount)
				if err != nil {
					b.Fatal(err)
				}

				for rows.Next() {
					err := rows.Scan(&br.ID, &br.ParentID, &br.Values)
					if err != nil {
						b.Fatal(err)
					}
				}

				if rows.Err() != nil {
					b.Fatal(rows.Err())
				}
			}
		})
	}
}
