/*
 * Copyright (C) 2015 - 2017 Wiky Lyu
 *
 * This program is free software: you can redistribute it and/or modify it
 * under the terms of the GNU General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 * See the GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.";
 */

package cipher

import (
	"crypto/aes"
	"crypto/cipher"
)

/* AES CFB模式 加密 */
type aesCFBEncrypter struct {
	stream cipher.Stream
}

func (e *aesCFBEncrypter) Encrypt(plain []byte) []byte {
	return cipherStreamXOR(e.stream, plain)
}

func newAESCFBEncrypter(key, iv []byte) Encrypter {
	block, _ := aes.NewCipher(key)

	stream := cipher.NewCFBEncrypter(block, iv)
	return &aesCFBEncrypter{stream}
}

/* AES CFB模式 解密 */
type aesCFBDecrypter struct {
	stream cipher.Stream
}

func (e *aesCFBDecrypter) Decrypt(encrypted []byte) []byte {
	return cipherStreamXOR(e.stream, encrypted)
}

func newAESCFBDecrypter(key, iv []byte) Decrypter {
	block, _ := aes.NewCipher(key)

	stream := cipher.NewCFBDecrypter(block, iv)
	return &aesCFBDecrypter{stream}
}
