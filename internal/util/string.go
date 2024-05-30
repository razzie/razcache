package util

import "unsafe"

func YoloString(p []byte) string {
	if len(p) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(p), len(p))
}

func YoloBytes(s string) []byte {
	if s == "" {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
