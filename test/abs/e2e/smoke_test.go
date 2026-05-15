//go:build abs_e2e

package e2e

import (
	"database/sql"
	"fmt"
	"testing"
)

type absSmokeCase struct {
	name              string
	instance          absInstance
	library           absLibrary
	dbParts           []string
	wantStoreMetadata bool
	wantItems         int
}

func TestABSHarnessSmokeResetContract(t *testing.T) {
	resetAndInitialScan(t)

	for _, tc := range []absSmokeCase{
		{
			name:              "plain_audiobooks",
			instance:          plainInstance,
			library:           audiobooksLibrary,
			dbParts:           []string{"test", "abs", "state", "plain", "config", "absdatabase.sqlite"},
			wantStoreMetadata: false,
			wantItems:         2,
		},
		{
			name:              "plain_books",
			instance:          plainInstance,
			library:           booksLibrary,
			dbParts:           []string{"test", "abs", "state", "plain", "config", "absdatabase.sqlite"},
			wantStoreMetadata: false,
			wantItems:         3,
		},
		{
			name:              "metadata_enabled_audiobooks",
			instance:          metadataEnabledInstance,
			library:           audiobooksLibrary,
			dbParts:           []string{"test", "abs", "state", "metadata-enabled", "config", "absdatabase.sqlite"},
			wantStoreMetadata: true,
			wantItems:         2,
		},
		{
			name:              "metadata_enabled_books",
			instance:          metadataEnabledInstance,
			library:           booksLibrary,
			dbParts:           []string{"test", "abs", "state", "metadata-enabled", "config", "absdatabase.sqlite"},
			wantStoreMetadata: true,
			wantItems:         3,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dbPath := pathFromRoot(tc.dbParts...)
			gotStoreMetadata := readStoreMetadataWithItem(t, dbPath)
			if gotStoreMetadata != tc.wantStoreMetadata {
				t.Fatalf(
					"storeMetadataWithItem mismatch for %s: got %t, want %t",
					tc.instance.name,
					gotStoreMetadata,
					tc.wantStoreMetadata,
				)
			}

			ctx := newABSScenarioContext(t, tc.instance, tc.library)
			waitForABSState(t, ctx, absStateExpectation{
				expectedCount: tc.wantItems,
				missingCount:  0,
			})
		})
	}
}

func readStoreMetadataWithItem(t *testing.T, dbPath string) bool {
	t.Helper()

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open ABS config database %s: %v", dbPath, err)
	}
	defer db.Close()

	var value int
	if err := db.QueryRow(`
		select json_extract(value, '$.storeMetadataWithItem')
		from settings
		where key = 'server-settings'
	`).Scan(&value); err != nil {
		t.Fatalf("read storeMetadataWithItem from %s: %v", dbPath, err)
	}

	switch value {
	case 0:
		return false
	case 1:
		return true
	default:
		t.Fatalf("unexpected storeMetadataWithItem value in %s: %s", dbPath, fmt.Sprint(value))
		return false
	}
}
