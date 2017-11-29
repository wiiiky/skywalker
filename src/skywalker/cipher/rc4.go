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
	_cipher "crypto/cipher"
	"crypto/md5"
	"crypto/rc4"
)

type rc4MD5Stream struct {
	stream _cipher.Stream
}

func (e *rc4MD5Stream) Encrypt(plain []byte) []byte {
	return cipherStreamXOR(e.stream, plain)
}

func (e *rc4MD5Stream) Decrypt(plain []byte) []byte {
	return cipherStreamXOR(e.stream, plain)
}

func newRC4MD5Stream(key, iv []byte) *rc4MD5Stream {
	h := md5.New()
	h.Write(key)
	h.Write(iv)
	rc4key := h.Sum(nil)

	stream, _ := rc4.NewCipher(rc4key)
	return &rc4MD5Stream{stream}
}

func newRC4MD5Encrypter(key, iv []byte) Encrypter {
	return newRC4MD5Stream(key, iv)
}

func newRC4MD5Decrypter(key, iv []byte) Decrypter {
	return newRC4MD5Stream(key, iv)
}
