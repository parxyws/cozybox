package util

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateOTP generates a random n-digit OTP
func GenerateRandomInteger(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("OTP length must be positive")
	}

	// Calculate the range: 10^(length-1) to 10^length - 1
	// For 6 digits: 100000 to 999999
	min := intPow(10, length-1)
	max := intPow(10, length) - 1

	// Generate random number in range [0, max-min]
	rangeSize := big.NewInt(int64(max - min + 1))
	randomNum, err := rand.Int(rand.Reader, rangeSize)
	if err != nil {
		return "", err
	}

	// Add min to get number in desired range
	otp := randomNum.Int64() + int64(min)

	// Format with leading zeros if needed
	return fmt.Sprintf("%0*d", length, otp), nil
}

// intPow calculates base^exp for integers
func intPow(base, exp int) int {
	result := 1
	for i := 0; i < exp; i++ {
		result *= base
	}
	return result
}
