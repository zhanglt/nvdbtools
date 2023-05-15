package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

var cveDBEncryptKey []byte = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

func getCVEDBEncryptKey() []byte {
	return cveDBEncryptKey
}

/**
* 函数用于使用 AES 对称加密算法加密给定的明文数据。
* 函数首先根据提供的密钥创建一个 AES 密码块（cipher），
* 然后使用密码块创建一个 GCM（Galois/Counter Mode）实例。
*接下来，函数生成一个随机的 nonce（用于加密和解密），并将其作为 AEAD（Authenticated Encryption with Additional Data）数据的一部分。
* 最后，函数使用 GCM 实例对明文数据进行加密操作，并返回加密后的密文数据。
**/
func encrypt(plaintext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	// 使用提供的密钥和加密模式创建的 GCM 对象进行加密操作
	// 随机生成一个 nonce，并将其作为附加的 AEAD（Authenticated Encryption with Additional Data）数据
	// 将明文数据进行加密，并返回加密后的密文数据
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

/**
* 该函数用于使用 AES 对称加密算法解密给定的密文数据。函数根据提供的密钥创建一个 AES 密码块（cipher），
* 使用密码块创建一个 GCM（Galois/Counter Mode）实例。
* 函数从密文中提取随机数（nonce）和实际的加密数据。
* 函数使用 GCM 实例对提取的数据进行解密操作，并返回解密后的明文数据。
**/
func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
