package grain_farmer

import (
	"common/middleware/db"
	"common/middleware/vipper"
	"common/utils"
	"os"
	"path/filepath"
	grainFarmerRepository "service/grain_farmer/repository"
	"strings"
	"testing"
)

func TestBackfillGrainFarmerSearchIndexes(t *testing.T) {
	if os.Getenv("GRAIN_FARMER_BACKFILL_INDEX") != "1" {
		t.Skip("set GRAIN_FARMER_BACKFILL_INDEX=1 to backfill grain_farmer search indexes")
	}
	setupGrainFarmerBackfillDB(t)

	key, err := requireFarmerCryptoKey()
	if err != nil {
		t.Fatalf("crypto key is required: %v", err)
	}

	repository := db.GetRepository[grainFarmerRepository.GrainFarmerRepository]()
	if repository == nil || repository.Db == nil {
		t.Fatal("database is not initialized")
	}
	if err := repository.EnsureTable(); err != nil {
		t.Fatalf("ensure grain_farmer table failed: %v", err)
	}

	const batchSize = 200
	lastID := 0
	updated := 0
	skipped := 0

	for {
		var farmers []*grainFarmerRepository.GrainFarmer
		err := repository.Db.
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(batchSize).
			Find(&farmers).Error
		if err != nil {
			t.Fatalf("load grain farmers failed after id %d: %v", lastID, err)
		}
		if len(farmers) == 0 {
			break
		}

		for _, farmer := range farmers {
			if farmer == nil {
				continue
			}
			lastID = farmer.Id

			name, idNumber, err := decryptFarmerBackfillValues(farmer, key)
			if err != nil {
				t.Fatalf("decrypt farmer id=%d failed: %v", farmer.Id, err)
			}
			if strings.TrimSpace(name) == "" && strings.TrimSpace(idNumber) == "" {
				skipped++
				continue
			}

			updates := map[string]interface{}{
				"name_digest":            utils.DigestField(name, key, farmerFieldScope("name")),
				"name_search":            farmerNamePrefixCodeWithKey(name, key),
				"id_number_digest":       utils.DigestField(idNumber, key, farmerFieldScope("id_number")),
				"id_number_last4_digest": farmerIDNumberLast4DigestWithKey(idNumber, key),
			}
			if err := repository.Db.Model(&grainFarmerRepository.GrainFarmer{}).
				Where("id = ?", farmer.Id).
				Updates(updates).Error; err != nil {
				t.Fatalf("backfill farmer id=%d failed: %v", farmer.Id, err)
			}
			updated++
		}
	}

	t.Logf("grain_farmer search index backfill completed: updated=%d skipped=%d", updated, skipped)
}

func decryptFarmerBackfillValues(farmer *grainFarmerRepository.GrainFarmer, key string) (string, string, error) {
	name, err := decryptFarmerField(farmer.Name, key, "name")
	if err != nil {
		return "", "", err
	}
	idNumber, err := decryptFarmerField(farmer.IDNumber, key, "id_number")
	if err != nil {
		return "", "", err
	}
	return strings.TrimSpace(name), strings.TrimSpace(idNumber), nil
}

func setupGrainFarmerBackfillDB(t *testing.T) {
	t.Helper()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory failed: %v", err)
	}
	configRoot := strings.TrimSpace(os.Getenv("GRAIN_FARMER_CONFIG_ROOT"))
	if configRoot == "" {
		configRoot = findGrainFarmerConfigRoot(t, originalWD)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
		if db.Db != nil {
			if sqlDB, err := db.Db.DB(); err == nil {
				_ = sqlDB.Close()
			}
		}
	})
	if err := os.Chdir(configRoot); err != nil {
		t.Fatalf("change working directory to %s failed: %v", configRoot, err)
	}
	vipper.Init()
	db.InitDB()
	if db.Db == nil {
		t.Fatal("database connection failed")
	}
}

func findGrainFarmerConfigRoot(t *testing.T, start string) string {
	t.Helper()
	for dir := start; ; dir = filepath.Dir(dir) {
		candidates := []string{
			filepath.Join(dir, "configs", "application.properties"),
			filepath.Join(dir, "server", "app-api", "configs", "application.properties"),
			filepath.Join(dir, "server", "manager-api", "configs", "application.properties"),
		}
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				return filepath.Dir(filepath.Dir(candidate))
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("configs/application.properties not found; set GRAIN_FARMER_CONFIG_ROOT to server/app-api or server/manager-api")
		}
	}
}
