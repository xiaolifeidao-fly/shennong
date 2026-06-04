package app_user

import (
	"common/middleware/vipper"
	"common/utils"
	"fmt"
	appUserRepository "service/app_user/repository"
	"strings"
)

const (
	appUserCryptoKeyConfig = "grain.app.user.crypto.key"
	appUserCryptoScope     = "app_user"
)

func appUserCryptoKey() string {
	return strings.TrimSpace(vipper.GetString(appUserCryptoKeyConfig))
}

func requireAppUserCryptoKey() (string, error) {
	key := appUserCryptoKey()
	if key == "" {
		return "", fmt.Errorf("%s is required", appUserCryptoKeyConfig)
	}
	return key, nil
}

func appUserFieldScope(field string) string {
	return appUserCryptoScope + ":" + field
}

func prepareAppUserIDCardForSave(entity *appUserRepository.AppUser) error {
	if entity == nil {
		return nil
	}
	key, err := requireAppUserCryptoKey()
	if err != nil {
		return err
	}
	plainIDNumber := strings.TrimSpace(entity.IDNumber)
	if plainIDNumber != "" && utils.IsEncryptedField(plainIDNumber) {
		plainIDNumber, err = utils.DecryptField(plainIDNumber, key, appUserFieldScope("id_number"))
		if err != nil {
			return err
		}
	}
	entity.IDNumberDigest = utils.DigestField(plainIDNumber, key, appUserFieldScope("id_number"))
	if entity.IDNumber, err = utils.EncryptField(plainIDNumber, key, appUserFieldScope("id_number")); err != nil {
		return err
	}
	return nil
}

func decryptAppUserIDNumber(entity *appUserRepository.AppUser) error {
	if entity == nil {
		return nil
	}
	key := appUserCryptoKey()
	if key == "" {
		return nil
	}
	var err error
	if entity.IDNumber, err = decryptAppUserField(entity.IDNumber, key, "id_number"); err != nil {
		return err
	}
	return nil
}

func decryptAppUserListRowIDNumber(row *appUserRepository.AppUserListRow) error {
	if row == nil {
		return nil
	}
	key := appUserCryptoKey()
	if key == "" {
		return nil
	}
	var err error
	if row.IDNumber, err = decryptAppUserField(row.IDNumber, key, "id_number"); err != nil {
		return err
	}
	return nil
}

func decryptAppUserField(value, key, field string) (string, error) {
	if strings.TrimSpace(value) == "" || !utils.IsEncryptedField(value) {
		return value, nil
	}
	return utils.DecryptField(value, key, appUserFieldScope(field))
}
