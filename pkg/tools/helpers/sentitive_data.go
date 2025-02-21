package helpers

// SensitiveData represents an encrypted string
type SensitiveData string

// NewSensitiveData creates a new encrypted sensitive data
func NewSensitiveData(plaintext string) (SensitiveData, error) {
	eh, err := GetEncryptionInstance()
	if err != nil {
		return "", err
	}

	encrypted, err := eh.EncryptBytes([]byte(plaintext))
	if err != nil {
		return "", err
	}

	return SensitiveData(encrypted), nil
}

// Decrypt decrypts the sensitive data
func (s SensitiveData) Decrypt() (string, error) {
	eh, err := GetEncryptionInstance()
	if err != nil {
		return "", err
	}

	data, err := eh.DecryptBytes(string(s))
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// String implements Stringer interface
func (s SensitiveData) String() string {
	return "[ENCRYPTED]"
}
