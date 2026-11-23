package files

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/Velarvo/velarvo-desktop/internal/apperr"
)

var ErrNotAFile = apperr.New("FILESYSTEM_NOT_A_FILE", "path is not a regular file")

const (
	PreviewText  = "text"
	PreviewImage = "image"
	PreviewNone  = "none"

	ReasonBinary   = "binary"
	ReasonTooLarge = "tooLarge"

	maxTextBytes  = 256 << 10
	maxImageBytes = 8 << 20
	sniffBytes    = 512
)

type Preview struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	MediaType string `json:"mediaType"`
	Text      string `json:"text,omitempty"`
	Data      []byte `json:"data,omitempty"`
	Truncated bool   `json:"truncated"`
	Reason    string `json:"reason,omitempty"`
}

func BuildPreview(name, path string, size int64, r io.Reader) (Preview, error) {
	preview := Preview{Kind: PreviewNone, Name: name, Path: path, Size: size}

	head := make([]byte, sniffBytes)
	n, err := io.ReadFull(r, head)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return Preview{}, fmt.Errorf("read file: %w", err)
	}
	head = head[:n]
	preview.MediaType = mediaType(name, head)

	if strings.HasPrefix(preview.MediaType, "image/") {
		if size > maxImageBytes {
			preview.Reason = ReasonTooLarge
			return preview, nil
		}
		data, truncated, err := readUpTo(head, r, maxImageBytes)
		if err != nil {
			return Preview{}, err
		}
		if truncated {
			preview.Reason = ReasonTooLarge
			return preview, nil
		}
		preview.Kind, preview.Data = PreviewImage, data
		return preview, nil
	}

	if !isText(head) {
		preview.Reason = ReasonBinary
		return preview, nil
	}
	data, truncated, err := readUpTo(head, r, maxTextBytes)
	if err != nil {
		return Preview{}, err
	}
	preview.Kind = PreviewText
	preview.Text = strings.ToValidUTF8(string(data), "�")
	preview.Truncated = truncated
	return preview, nil
}

func readUpTo(head []byte, r io.Reader, limit int) ([]byte, bool, error) {
	if len(head) >= limit {
		return head[:limit], true, nil
	}
	rest, err := io.ReadAll(io.LimitReader(r, int64(limit-len(head))+1))
	if err != nil {
		return nil, false, fmt.Errorf("read file: %w", err)
	}
	data := append(head, rest...) //nolint:gocritic // head is ours to extend
	if len(data) > limit {
		return data[:limit], true, nil
	}
	return data, false, nil
}

func mediaType(name string, head []byte) string {
	value := mime.TypeByExtension(filepath.Ext(name))
	if value == "" {
		if len(head) == 0 {
			return "text/plain"
		}
		value = http.DetectContentType(head)
	}
	base, _, _ := strings.Cut(value, ";")
	return strings.ToLower(strings.TrimSpace(base))
}

func isText(head []byte) bool {
	if bytes.IndexByte(head, 0) >= 0 {
		return false
	}

	suspicious := 0
	for i := 0; i < len(head); {
		r, width := utf8.DecodeRune(head[i:])
		if r == utf8.RuneError && width == 1 {
			if i+utf8.UTFMax > len(head) {
				break
			}
			suspicious++
		} else if r < 0x20 && !strings.ContainsRune("\t\n\r\f\v", r) {
			suspicious++
		}
		i += width
	}
	return suspicious*10 <= len(head)
}
