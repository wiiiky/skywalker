/*
 * Copyright (C) 2015 Wiky L
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
    "skywalker/cipher/chacha20"
)

type chacha20Stream struct {
    stream _cipher.Stream
}

func (s *chacha20Stream) Encrypt(data []byte) []byte {
    return cipherStreamXOR(s.stream, data)
}

func (s *chacha20Stream) Decrypt(data []byte) []byte {
    return cipherStreamXOR(s.stream, data)
}

func newChacha20Stream(key, iv []byte) *chacha20Stream{
	stream, _ := chacha20.New(key, iv)
	return &chacha20Stream{stream}
}

func newChacha20Encrypter(key, iv []byte) Encrypter {
    return newChacha20Stream(key, iv)
}

func newChacha20Decrypter(key, iv []byte) Decrypter {
    return newChacha20Stream(key, iv)
}
