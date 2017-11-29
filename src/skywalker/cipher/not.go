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

/*
 * 取反
 */

type notEncrypter struct {
}

func (e *notEncrypter) Encrypt(plain []byte) []byte {
	for i, v := range plain {
		plain[i] = ^v
	}
	return plain
}

func newNotEncrypter(key, iv []byte) Encrypter {
	return &notEncrypter{}
}

type notDecrypter struct {
}

func (e *notDecrypter) Decrypt(encrypted []byte) []byte {
	for i, v := range encrypted {
		encrypted[i] = ^v
	}
	return encrypted
}

func newNotDecrypter(key, iv []byte) Decrypter {
	return &noneDecrypter{}
}
