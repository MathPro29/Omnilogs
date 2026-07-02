package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

// intPtrFromAny แปลงค่าจาก any (มักจะเป็น float64 เวลา unmarshal JSON) ให้เป็น pointer ของ int
func intPtrFromAny(value any) *int {
	switch typed := value.(type) {
	case float64:
		result := int(typed)
		return &result
	case int:
		result := typed
		return &result
	}
	return nil
}

// stringPtrFromAny แปลงค่าจาก any ให้เป็น pointer ของ string พร้อมลบช่องว่างหัวท้าย
func stringPtrFromAny(value any) *string {
	typed, ok := value.(string)
	if !ok {
		return nil
	}
	typed = strings.TrimSpace(typed)
	if typed == "" {
		return nil
	}
	return &typed
}

// timePtrFromAny แปลงค่าจาก any ให้เป็น time.Time โดยคาดหวังรูปแบบ RFC3339
func timePtrFromAny(value any) (time.Time, bool) {
	raw, ok := value.(string)
	if !ok {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

// normalizeIndexSegment จัดการข้อความให้ปลอดภัยสำหรับนำไปตั้งชื่อ Index ใน Elasticsearch
// โดยแปลงเป็นตัวพิมพ์เล็ก และเปลี่ยน _ หรือช่องว่างให้เป็น - (Hyphen)
func normalizeIndexSegment(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, " ", "-")
	if value == "" {
		return "omnilogs"
	}
	return value
}

// newUUID สุ่มสร้างรหัส UUID v4 แบบง่ายๆ โดยไม่ต้องพึ่งพาไลบรารีภายนอก
func newUUID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "00000000-0000-4000-8000-000000000000"
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	dst := make([]byte, 36)
	hex.Encode(dst[0:8], bytes[0:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], bytes[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], bytes[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], bytes[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:36], bytes[10:16])
	return string(dst)
}
