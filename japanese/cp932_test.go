package japanese_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"unicode/utf8"

	"github.com/akihiroy/pytextcodec/japanese"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/text/transform"
)

// TestData structures for MessagePack files
type (
	EncoderTestData map[string][]byte
	DecoderTestData map[string]string
)

// loadEncoderTestData loads encoder test data from MessagePack file
func loadEncoderTestData(t *testing.T) EncoderTestData {
	// Get the directory where this test file is located
	testDir := filepath.Dir(".")
	testDataDir := filepath.Join(testDir, "testdata")

	// Load encoder test data
	encoderFile := filepath.Join(testDataDir, "cp932_encoder.msgpack")
	encoderData, err := os.ReadFile(encoderFile)
	if err != nil {
		t.Fatalf("Failed to read encoder test data: %v", err)
	}

	var encoderTestData EncoderTestData
	if err := msgpack.Unmarshal(encoderData, &encoderTestData); err != nil {
		t.Fatalf("Failed to unmarshal encoder test data: %v", err)
	}

	return encoderTestData
}

// loadDecoderTestData loads decoder test data from MessagePack file
func loadDecoderTestData(t *testing.T) DecoderTestData {
	// Get the directory where this test file is located
	testDir := filepath.Dir(".")
	testDataDir := filepath.Join(testDir, "testdata")

	// Load decoder test data
	decoderFile := filepath.Join(testDataDir, "cp932_decoder.msgpack")
	decoderData, err := os.ReadFile(decoderFile)
	if err != nil {
		t.Fatalf("Failed to read decoder test data: %v", err)
	}

	var decoderTestData DecoderTestData
	if err := msgpack.Unmarshal(decoderData, &decoderTestData); err != nil {
		t.Fatalf("Failed to unmarshal decoder test data: %v", err)
	}

	return decoderTestData
}

// TestCP932EncoderWithTestData tests encoder using comprehensive Unicode ranges
func TestCP932EncoderWithTestData(t *testing.T) {
	t.Parallel()
	encoderTestData := loadEncoderTestData(t)

	// Prepare test cases for Unicode ranges
	var testCases []struct {
		name       string
		startRange uint32
		endRange   uint32
	}

	// Create test cases for each upper word (0x000000-0x10FFFF)
	for upperWord := uint32(0); upperWord <= 0x10; upperWord++ {
		startRange := upperWord << 16
		endRange := startRange + 0xFFFF
		if upperWord == 0x10 {
			endRange = 0x10FFFF // Last range ends at 0x10FFFF
		}

		testCases = append(testCases, struct {
			name       string
			startRange uint32
			endRange   uint32
		}{
			name:       fmt.Sprintf("range_0x%05X_0x%05X", startRange, endRange),
			startRange: startRange,
			endRange:   endRange,
		})
	}

	// Execute all test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			encoder := japanese.CP932.NewEncoder()

			// Test each code point in the range
			for code := tc.startRange; code <= tc.endRange; code++ {
				codeStr := fmt.Sprintf("U+%06X", code)
				expectedBytes, exists := encoderTestData[codeStr]

				// Convert code point to character
				char := rune(code)
				runeLen := utf8.RuneLen(char)
				if runeLen == -1 {
					require.False(t, exists, "Invalid code point %s should not exist in test data", codeStr)
					continue
				}
				input := make([]byte, runeLen)
				utf8.EncodeRune(input, char)

				// Encode using our CP932 implementation
				result, _, err := transform.Bytes(encoder, input)
				if exists {
					require.NoError(t, err, "Encoding failed for %s. Expected: 0x%x", codeStr, expectedBytes)
					// Compare with expected result
					require.Equal(t, expectedBytes, result, "Encoding mismatch for %s", codeStr)
				} else {
					require.Error(t, err, "Encoding should fail for %s", codeStr)
				}

			}
		})
	}
}

// TestCP932DecoderWithTestData tests decoder using comprehensive byte combinations
func TestCP932DecoderWithTestData(t *testing.T) {
	t.Parallel()
	decoderTestData := loadDecoderTestData(t)

	// Prepare test cases for each c0 value (0x00-0xFF)
	var testCases []struct {
		name       string
		c0         int
		singleByte bool
	}

	for c0 := 0x00; c0 <= 0xFF; c0++ {
		if isSingleByteRange(c0) {
			// Single-byte character
			testCases = append(testCases, struct {
				name       string
				c0         int
				singleByte bool
			}{
				name:       fmt.Sprintf("0x%02X", c0),
				c0:         c0,
				singleByte: true, // single-byte
			})
		} else {
			// 2-byte character
			testCases = append(testCases, struct {
				name       string
				c0         int
				singleByte bool
			}{
				name:       fmt.Sprintf("0x%02X", c0),
				c0:         c0,
				singleByte: false, // double-byte
			})
		}
	}

	// Execute all test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			decoder := japanese.CP932.NewDecoder()

			if tc.singleByte {
				// Test single-byte decoding
				inputBytes := []byte{byte(tc.c0)}
				expectedCodeStr, exists := decoderTestData[fmt.Sprintf("%02X", tc.c0)]

				if !exists {
					// This byte is not valid as single-byte in CP932
					// Expected to return 0xFFFD (REPLACEMENT CHARACTER)
					expectedCodeStr = "U+00FFFD"
				}

				// Test decoding result
				testDecodingResult(t, decoder, inputBytes, expectedCodeStr, fmt.Sprintf("single-byte 0x%02X", tc.c0))
			} else {
				// Test all 2-byte combinations for this c0 value
				for c1 := 0x00; c1 <= 0xFF; c1++ {
					inputBytes := []byte{byte(tc.c0), byte(c1)}
					expectedCodeStr, exists := decoderTestData[fmt.Sprintf("%02X%02X", tc.c0, c1)]

					if !exists {
						// This byte combination is not valid in CP932
						// Expected to return 0xFFFD (REPLACEMENT CHARACTER)
						expectedCodeStr = "U+00FFFD"
					}

					// Test decoding result
					testDecodingResult(t, decoder, inputBytes, expectedCodeStr, fmt.Sprintf("0x%02X%02X", tc.c0, c1))
				}
			}
		})
	}
}

// testDecodingResult is a helper function to test decoding results
func testDecodingResult(
	t *testing.T,
	decoder transform.Transformer,
	inputBytes []byte,
	expectedCodeStr, testName string,
) {
	result, _, err := transform.Bytes(decoder, inputBytes)
	assert.NoError(t, err, "Decoding failed for valid %s", testName)
	assert.NotEmpty(t, result, "Decoding result is empty for valid %s", testName)

	runes := []rune(string(result))
	assert.NotEmpty(t, runes, "No runes in decoding result for %s", testName)

	actualCodeStr := fmt.Sprintf("U+%06X", runes[0])
	assert.Equal(t, expectedCodeStr, actualCodeStr, "Decoding mismatch for %s", testName)
}

// isSingleByteRange checks if the byte value corresponds to single-byte character range
func isSingleByteRange(c0 int) bool {
	return c0 <= 0x80 || (0xA0 <= c0 && c0 <= 0xDF) || (0xFD <= c0 && c0 <= 0xFF)
}

func TestCP932Encoding(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected []byte
	}{
		{
			name:     "ASCII characters",
			input:    "Hello, World!",
			expected: []byte("Hello, World!"),
		},
		{
			name:     "Japanese characters",
			input:    "こんにちは〜", // includes wave dash (U+301C)
			expected: []byte{0x82, 0xB1, 0x82, 0xF1, 0x82, 0xC9, 0x82, 0xBF, 0x82, 0xCD, 0x81, 0x60},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := japanese.CP932.NewEncoder()
			result, _, err := transform.Bytes(encoder, []byte(tt.input))
			assert.NoError(t, err, "Encoding error")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCP932EncodingBufferHandling(t *testing.T) {
	t.Parallel()
	t.Run("ErrShortDst1", func(t *testing.T) {
		t.Parallel()
		encoder := japanese.CP932.NewEncoder()
		result, _, err := transform.Append(encoder, make([]byte, 0, 1), []byte("あ"))
		assert.NoError(t, err, "Encoding error")
		assert.Equal(t, []byte{0x82, 0xA0}, result)
	})
	t.Run("ErrShortDst2", func(t *testing.T) {
		t.Parallel()
		encoder := japanese.CP932.NewEncoder()
		result, _, err := transform.Append(encoder, make([]byte, 0, 1), []byte("\u00A2"))
		assert.NoError(t, err, "Encoding error")
		assert.Equal(t, []byte{0x81, 0x91}, result)
	})
	t.Run("ErrShortSrc", func(t *testing.T) {
		t.Parallel()
		encoder := japanese.CP932.NewEncoder()
		buf := bytes.NewBuffer([]byte{})
		writer := transform.NewWriter(buf, encoder)
		n, err := writer.Write([]byte{0xC2})
		assert.NoError(t, err, "Encoding error")
		assert.Equal(t, 1, n)
		assert.Equal(t, 0, buf.Len())
		n, err = writer.Write([]byte{0x80})
		assert.NoError(t, err, "Encoding error")
		assert.Equal(t, 1, n)
		assert.Equal(t, []byte{0x80}, buf.Bytes())
	})
}

func TestCP932Decoding(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "ASCII characters",
			input:    []byte("Hello, World!"),
			expected: "Hello, World!",
		},
		{
			name:     "Japanese characters",
			input:    []byte{0x82, 0xB1, 0x82, 0xF1, 0x82, 0xC9, 0x82, 0xBF, 0x82, 0xCD},
			expected: "こんにちは",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := japanese.CP932.NewDecoder()
			result, _, err := transform.Bytes(decoder, tt.input)
			assert.NoError(t, err, "Decoding error")
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

func TestCP932DecoderBufferHandling(t *testing.T) {
	t.Parallel()
	t.Run("ErrShortSrc1", func(t *testing.T) {
		t.Parallel()
		decoder := japanese.CP932.NewDecoder()
		buf := bytes.NewBuffer([]byte{})
		writer := transform.NewWriter(buf, decoder)
		n, err := writer.Write([]byte{0x82})
		assert.NoError(t, err, "Decoding error")
		assert.Equal(t, 1, n)
		assert.Equal(t, 0, buf.Len())
		n, err = writer.Write([]byte{0xA0})
		assert.NoError(t, err, "Decoding error")
		assert.Equal(t, 1, n)
		assert.Equal(t, "あ", buf.String())
	})
	t.Run("ErrShortSrc2", func(t *testing.T) {
		t.Parallel()
		decoder := japanese.CP932.NewDecoder()
		buf := bytes.NewBuffer([]byte{})
		writer := transform.NewWriter(buf, decoder)
		n, err := writer.Write([]byte{0xF0})
		assert.NoError(t, err, "Decoding error")
		assert.Equal(t, 1, n)
		assert.Equal(t, 0, buf.Len())
		n, err = writer.Write([]byte{0x40})
		assert.NoError(t, err, "Decoding error")
		assert.Equal(t, 1, n)
		assert.Equal(t, []byte{0xEE, 0x80, 0x80}, buf.Bytes())
	})
	t.Run("ErrShortDst", func(t *testing.T) {
		t.Parallel()
		decoder := japanese.CP932.NewDecoder()
		result, _, err := transform.Append(decoder, make([]byte, 0, 1), []byte{0xA0})
		assert.NoError(t, err, "Decoding error")
		assert.Equal(t, []byte{0xEF, 0xA3, 0xB0}, result)
	})
}

func TestCP932RoundTrip(t *testing.T) {
	testStrings := []string{
		"Hello, World!",
		"こんにちは～", // includes full-width tilde (U+FF5E)
		"Hello, 世界!",
		"Programmingプログラミングﾌﾟﾛｸﾞﾗﾐﾝｸﾞ",
	}

	for _, input := range testStrings {
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			// Encode
			encoder := japanese.CP932.NewEncoder()
			encoded, _, err := transform.Bytes(encoder, []byte(input))
			assert.NoError(t, err, "Encoding error")

			// Decode
			decoder := japanese.CP932.NewDecoder()
			decoded, _, err := transform.Bytes(decoder, encoded)
			assert.NoError(t, err, "Decoding error")

			assert.Equal(t, input, string(decoded), "Round trip failed")
		})
	}
}
