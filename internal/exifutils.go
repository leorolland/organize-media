package internal

import (
	"fmt"
	"os"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

// GetExifDate retrieves the EXIF date from a file using the provided decode and dateTime functions.
func GetExifDate(path string) (time.Time, error) {
	file, err := os.Open(path)
	if err != nil {
		return time.Time{}, err
	}
	defer file.Close()

	// Parse EXIF metadata
	exifData, err := exif.Decode(file)
	if err != nil {
		return time.Time{}, err
	}

	// Extract DateTime field
	date, err := exifData.DateTime()
	if err != nil {
		return time.Time{}, err
	}

	return date, nil
}

func extractEXIFBlock(data []byte) []byte {
	// Search for APP1 EXIF marker (0xFFE1)
	for i := 0; i < len(data)-1; i++ {
		if data[i] == 0xFF && data[i+1] == 0xE1 { // APP1 marker
			// Read segment length (2 bytes after the marker)
			if i+4 > len(data) {
				break
			}
			segLen := int(data[i+2])<<8 + int(data[i+3])

			if i+2+segLen > len(data) {
				break
			}
			return data[i : i+2+segLen]
		}
	}
	return nil
}

func writeJPEGWithEXIF(dest string, exifBlock []byte, encodedImage []byte) error {
	// Create destination file
	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create %s: %v", dest, err)
	}
	defer out.Close()

	// Write JPEG SOI header (0xFFD8)
	if _, err := out.Write([]byte{0xFF, 0xD8}); err != nil {
		return fmt.Errorf("failed to write SOI header: %v", err)
	}

	// Write EXIF block if available
	if exifBlock != nil {
		if _, err := out.Write(exifBlock); err != nil {
			return fmt.Errorf("failed to write EXIF block: %v", err)
		}
	}

	// Skip SOI (first 2 bytes) from the encoded result and append the rest
	if _, err := out.Write(encodedImage[2:]); err != nil {
		return fmt.Errorf("failed to write encoded image: %v", err)
	}

	return nil
}
