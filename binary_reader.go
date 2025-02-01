// ./binary_reader.go
package jpleph

/*
Package jpleph provides helper functions for reading binary data.

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program; if not, write to the Free Software
Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA
02110-1301, USA.

Authorship:
Mohammad Shafiee authored this Go code as a translation of the original C code.
The C version was a translation of Fortran-77 code originally written by
Piotr A. Dybczynski and later revised by Bill J Gray.
*/

import (
	"encoding/binary"
	"io"
	"math"
)

// defaultByteOrder specifies the default byte order for reading binary data.
// JPL ephemeris files are typically in little-endian format.
var defaultByteOrder = binary.LittleEndian

// byteOrder is a configurable byte order for reading binary data.
// Defaults to little-endian but can be changed if needed.
var byteOrder binary.ByteOrder = defaultByteOrder

// SetByteOrder allows changing the byte order for reading binary data.
// Use binary.LittleEndian or binary.BigEndian.
func SetByteOrder(order binary.ByteOrder) {
	byteOrder = order
}

// getNumber reads a value of the specified type from the io.Reader using the configured byte order.
// It takes an io.Reader and a pointer to the variable where the read value will be stored.
// Returns an error if reading fails.
func getNumber(r io.Reader, data any) error {
	return binary.Read(r, byteOrder, data)
}

// getUint16 reads a uint16 value in the configured byte order.
func getUint16(r io.Reader) (uint16, error) {
	var val uint16
	err := getNumber(r, &val)
	return val, err
}

// getUint32 reads a uint32 value in the configured byte order.
func getUint32(r io.Reader) (uint32, error) {
	var val uint32
	err := getNumber(r, &val)
	return val, err
}

// getUint64 reads a uint64 value in the configured byte order.
func getUint64(r io.Reader) (uint64, error) {
	var val uint64
	err := getNumber(r, &val)
	return val, err
}

// getInt16 reads an int16 value in the configured byte order.
func getInt16(r io.Reader) (int16, error) {
	var val int16
	err := getNumber(r, &val)
	return val, err
}

// getInt32 reads an int32 value in the configured byte order.
func getInt32(r io.Reader) (int32, error) {
	var val int32
	err := getNumber(r, &val)
	return val, err
}

// getInt64 reads an int64 value in the configured byte order.
func getInt64(r io.Reader) (int64, error) {
	var val int64
	err := getNumber(r, &val)
	return val, err
}

// getFloat64 reads a float64 (double-precision) value in the configured byte order.
func getFloat64(r io.Reader) (float64, error) {
	var val float64
	err := getNumber(r, &val)
	return val, err
}

// uInt32FromBytes converts a byte slice to a uint32 value using the configured byte order.
func uInt32FromBytes(b []byte) uint32 {
	return byteOrder.Uint32(b)
}

// float64FromBytes converts a byte slice to a float64 value using the configured byte order.
func float64FromBytes(b []byte) float64 {
	return math.Float64frombits(byteOrder.Uint64(b))
}

// swapBytes32 performs in-place byte swapping for a 32-bit unsigned integer.
// Useful for handling ephemeris files with different byte orders.
func swapBytes32(val *uint32) {
	b := make([]byte, 4)
	byteOrder.PutUint32(b, *val)

	// Swap bytes: 0 <-> 3, 1 <-> 2
	b[0], b[3] = b[3], b[0]
	b[1], b[2] = b[2], b[1]

	*val = byteOrder.Uint32(b)
}

// swapBytes64 performs in-place byte swapping for a 64-bit floating-point number.
// Useful for handling ephemeris files with different byte orders.
func swapBytes64(val *float64) {
	b := make([]byte, 8)
	byteOrder.PutUint64(b, math.Float64bits(*val)) // Convert float64 to uint64 bits for byte manipulation

	// Swap bytes: 0 <-> 7, 1 <-> 6, 2 <-> 5, 3 <-> 4
	b[0], b[7] = b[7], b[0]
	b[1], b[6] = b[6], b[1]
	b[2], b[5] = b[5], b[2]
	b[3], b[4] = b[4], b[3]

	*val = math.Float64frombits(byteOrder.Uint64(b)) // Interpret swapped bytes as float64
}

// swapBytes64Slice applies SwapBytes64 to each element in a float64 slice.
func swapBytes64Slice(slice []float64) {
	for i := range slice {
		swapBytes64(&slice[i]) // Byte-swap each float64 value in the slice
	}
}
