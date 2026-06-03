package grain_farmer

import (
	"common/middleware/vipper"
	"common/utils"
	"fmt"
	grainFarmerDTO "service/grain_farmer/dto"
	grainFarmerRepository "service/grain_farmer/repository"
	"strings"
)

const (
	farmerCryptoKeyConfig = "grain.farmer.crypto.key"
	farmerCryptoScope     = "grain_farmer"
)

func farmerCryptoKey() string {
	return strings.TrimSpace(vipper.GetString(farmerCryptoKeyConfig))
}

func requireFarmerCryptoKey() (string, error) {
	key := farmerCryptoKey()
	if key == "" {
		return "", fmt.Errorf("%s is required", farmerCryptoKeyConfig)
	}
	return key, nil
}

func prepareFarmerForSave(entity *grainFarmerRepository.GrainFarmer) error {
	if entity == nil {
		return nil
	}
	key, err := requireFarmerCryptoKey()
	if err != nil {
		return err
	}
	plainIDNumber := strings.TrimSpace(entity.IDNumber)
	if plainIDNumber != "" && utils.IsEncryptedField(plainIDNumber) {
		plainIDNumber, err = utils.DecryptField(plainIDNumber, key, farmerFieldScope("id_number"))
		if err != nil {
			return err
		}
	}
	entity.IDNumberDigest = utils.DigestField(plainIDNumber, key, farmerFieldScope("id_number"))
	if entity.Name, err = utils.EncryptField(entity.Name, key, farmerFieldScope("name")); err != nil {
		return err
	}
	if entity.IDNumber, err = utils.EncryptField(plainIDNumber, key, farmerFieldScope("id_number")); err != nil {
		return err
	}
	if entity.BankNumber, err = utils.EncryptField(entity.BankNumber, key, farmerFieldScope("bank_number")); err != nil {
		return err
	}
	if entity.BankName, err = utils.EncryptField(entity.BankName, key, farmerFieldScope("bank_name")); err != nil {
		return err
	}
	return nil
}

func decryptFarmerEntity(entity *grainFarmerRepository.GrainFarmer) error {
	if entity == nil {
		return nil
	}
	key := farmerCryptoKey()
	if key == "" {
		return nil
	}
	var err error
	if entity.Name, err = decryptFarmerField(entity.Name, key, "name"); err != nil {
		return err
	}
	if entity.IDNumber, err = decryptFarmerField(entity.IDNumber, key, "id_number"); err != nil {
		return err
	}
	if entity.Phone, err = decryptFarmerField(entity.Phone, key, "phone"); err != nil {
		return err
	}
	if entity.Address, err = decryptFarmerField(entity.Address, key, "address"); err != nil {
		return err
	}
	if entity.BankNumber, err = decryptFarmerField(entity.BankNumber, key, "bank_number"); err != nil {
		return err
	}
	if entity.BankName, err = decryptFarmerField(entity.BankName, key, "bank_name"); err != nil {
		return err
	}
	return nil
}

func decryptFarmerEntities(entities []*grainFarmerRepository.GrainFarmer) error {
	for _, entity := range entities {
		if err := decryptFarmerEntity(entity); err != nil {
			return err
		}
	}
	return nil
}

func decryptFarmerListRows(rows []*grainFarmerRepository.GrainFarmerListRow) error {
	key := farmerCryptoKey()
	if key == "" {
		return nil
	}
	for _, row := range rows {
		var err error
		if row.Name, err = decryptFarmerField(row.Name, key, "name"); err != nil {
			return err
		}
		if row.IDNumber, err = decryptFarmerField(row.IDNumber, key, "id_number"); err != nil {
			return err
		}
		if row.Phone, err = decryptFarmerField(row.Phone, key, "phone"); err != nil {
			return err
		}
		if row.Address, err = decryptFarmerField(row.Address, key, "address"); err != nil {
			return err
		}
		if row.BankNumber, err = decryptFarmerField(row.BankNumber, key, "bank_number"); err != nil {
			return err
		}
		if row.BankName, err = decryptFarmerField(row.BankName, key, "bank_name"); err != nil {
			return err
		}
	}
	return nil
}

func decryptFarmerDTO(dto *grainFarmerDTO.GrainFarmerDTO) error {
	if dto == nil {
		return nil
	}
	key := farmerCryptoKey()
	if key == "" {
		return nil
	}
	var err error
	if dto.Name, err = decryptFarmerField(dto.Name, key, "name"); err != nil {
		return err
	}
	if dto.IDNumber, err = decryptFarmerField(dto.IDNumber, key, "id_number"); err != nil {
		return err
	}
	if dto.Phone, err = decryptFarmerField(dto.Phone, key, "phone"); err != nil {
		return err
	}
	if dto.Address, err = decryptFarmerField(dto.Address, key, "address"); err != nil {
		return err
	}
	if dto.BankNumber, err = decryptFarmerField(dto.BankNumber, key, "bank_number"); err != nil {
		return err
	}
	if dto.BankName, err = decryptFarmerField(dto.BankName, key, "bank_name"); err != nil {
		return err
	}
	return nil
}

func decryptFarmerField(value, key, field string) (string, error) {
	if strings.TrimSpace(value) == "" || !utils.IsEncryptedField(value) {
		return value, nil
	}
	return utils.DecryptField(value, key, farmerFieldScope(field))
}

func DecryptFarmerSensitiveValues(name, idNumber, bankNumber, bankName string) (string, string, string, string, error) {
	name, idNumber, _, _, bankNumber, bankName, err := DecryptFarmerProfileValues(name, idNumber, "", "", bankNumber, bankName)
	return name, idNumber, bankNumber, bankName, err
}

func DecryptFarmerProfileValues(name, idNumber, phone, address, bankNumber, bankName string) (string, string, string, string, string, string, error) {
	key := farmerCryptoKey()
	if key == "" {
		return name, idNumber, phone, address, bankNumber, bankName, nil
	}
	var err error
	if name, err = decryptFarmerField(name, key, "name"); err != nil {
		return "", "", "", "", "", "", err
	}
	if idNumber, err = decryptFarmerField(idNumber, key, "id_number"); err != nil {
		return "", "", "", "", "", "", err
	}
	if phone, err = decryptFarmerField(phone, key, "phone"); err != nil {
		return "", "", "", "", "", "", err
	}
	if address, err = decryptFarmerField(address, key, "address"); err != nil {
		return "", "", "", "", "", "", err
	}
	if bankNumber, err = decryptFarmerField(bankNumber, key, "bank_number"); err != nil {
		return "", "", "", "", "", "", err
	}
	if bankName, err = decryptFarmerField(bankName, key, "bank_name"); err != nil {
		return "", "", "", "", "", "", err
	}
	return name, idNumber, phone, address, bankNumber, bankName, nil
}

func farmerIDNumberDigest(value string) (string, error) {
	key, err := requireFarmerCryptoKey()
	if err != nil {
		return "", err
	}
	return utils.DigestField(value, key, farmerFieldScope("id_number")), nil
}

func farmerFieldScope(field string) string {
	return farmerCryptoScope + ":" + field
}
