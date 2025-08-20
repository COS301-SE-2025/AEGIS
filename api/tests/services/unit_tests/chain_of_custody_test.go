package unit_tests

import (
	"context"
	"testing"
	"time"

	"aegis-api/services_/chain_of_custody"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupChainOfCustodyTestDB(t *testing.T) *gorm.DB {
	dbName := "file:" + uuid.New().String() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&chain_of_custody.ChainOfCustody{})
	require.NoError(t, err)
	return db
}

func TestAddAndGetEntry(t *testing.T) {
	db := setupChainOfCustodyTestDB(t)
	repo := chain_of_custody.NewChainOfCustodyRepository(db)
	service := chain_of_custody.NewChainOfCustodyService(repo)

	acqDate := time.Now().UTC()
	custody := &chain_of_custody.ChainOfCustody{
		EvidenceID:      uuid.New(),
		Custodian:       "John Doe",
		AcquisitionDate: &acqDate,
		AcquisitionTool: "ToolX",
		SystemInfo:      datatypes.JSON([]byte(`{"os":"Linux"}`)),
		ForensicInfo:    datatypes.JSON([]byte(`{"hash":"abc123"}`)),
	}
	err := service.AddEntry(context.Background(), custody)
	require.NoError(t, err)

	fetched, err := service.GetEntry(context.Background(), custody.ID)
	require.NoError(t, err)
	require.Equal(t, custody.Custodian, fetched.Custodian)
	require.Equal(t, custody.EvidenceID, fetched.EvidenceID)
}

func TestUpdateEntry(t *testing.T) {
	db := setupChainOfCustodyTestDB(t)
	repo := chain_of_custody.NewChainOfCustodyRepository(db)
	service := chain_of_custody.NewChainOfCustodyService(repo)

	acqDate := time.Now().UTC()
	custody := &chain_of_custody.ChainOfCustody{
		EvidenceID:      uuid.New(),
		Custodian:       "Jane Doe",
		AcquisitionDate: &acqDate,
		AcquisitionTool: "ToolY",
		SystemInfo:      datatypes.JSON([]byte(`{"os":"Windows"}`)),
		ForensicInfo:    datatypes.JSON([]byte(`{"hash":"def456"}`)),
	}
	err := service.AddEntry(context.Background(), custody)
	require.NoError(t, err)

	custody.Custodian = "Jane Smith"
	err = service.UpdateEntry(context.Background(), custody)
	require.NoError(t, err)

	fetched, err := service.GetEntry(context.Background(), custody.ID)
	require.NoError(t, err)
	require.Equal(t, "Jane Smith", fetched.Custodian)
}

func TestGetEntriesByEvidenceID(t *testing.T) {
	db := setupChainOfCustodyTestDB(t)
	repo := chain_of_custody.NewChainOfCustodyRepository(db)
	service := chain_of_custody.NewChainOfCustodyService(repo)

	evidenceID := uuid.New()
	acqDate := time.Now().UTC()
	custody1 := &chain_of_custody.ChainOfCustody{
		EvidenceID:      evidenceID,
		Custodian:       "Custodian1",
		AcquisitionDate: &acqDate,
		AcquisitionTool: "ToolA",
		SystemInfo:      datatypes.JSON([]byte(`{"os":"Linux"}`)),
		ForensicInfo:    datatypes.JSON([]byte(`{"hash":"hash1"}`)),
	}
	custody2 := &chain_of_custody.ChainOfCustody{
		EvidenceID:      evidenceID,
		Custodian:       "Custodian2",
		AcquisitionDate: &acqDate,
		AcquisitionTool: "ToolB",
		SystemInfo:      datatypes.JSON([]byte(`{"os":"Linux"}`)),
		ForensicInfo:    datatypes.JSON([]byte(`{"hash":"hash2"}`)),
	}
	err := service.AddEntry(context.Background(), custody1)
	require.NoError(t, err)
	err = service.AddEntry(context.Background(), custody2)
	require.NoError(t, err)

	entries, err := service.GetEntries(context.Background(), evidenceID)
	require.NoError(t, err)
	require.Len(t, entries, 2)
}
