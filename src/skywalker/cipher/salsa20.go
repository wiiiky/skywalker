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
	"skywalker/cipher/salsa20"
)

type salsaStream struct {
	nonce   []byte
	key     [32]byte
	counter int
}

func (stream *salsaStream) XORKeyStream(out, in []byte) {
	key := stream.key
	nonce := stream.nonce

	salsa20.XORKeyStream(out, in, nonce, &key)
}

func (stream *salsaStream) Encrypt(data []byte) []byte {
	return cipherStreamXOR(stream, data)
}

func (stream *salsaStream) Decrypt(data []byte) []byte {
	return cipherStreamXOR(stream, data)
}

func newSalsa20Stream(key, iv []byte) *salsaStream {
	var key32 [32]byte
	copy(key32[:], key)
	return &salsaStream{iv, key32, 0}
}

func newSalsa20Encrypter(key, iv []byte) Encrypter {
	return newSalsa20Stream(key, iv)
}

func newSalsa20Decrypter(key, iv []byte) Decrypter {
	return newSalsa20Stream(key, iv)
}
