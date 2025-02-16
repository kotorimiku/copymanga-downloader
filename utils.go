package main

import (
	"bytes"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

const passwordChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"

func GeneratePassword(length int) string {
	randSource := rand.NewSource(time.Now().UnixNano())
	randGen := rand.New(randSource)

	password := make([]byte, length)
	for i := range password {
		password[i] = passwordChars[randGen.Intn(len(passwordChars))]
	}
	return string(password)
}

const usernameChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateUsername(length int) string {
	randSource := rand.NewSource(time.Now().UnixNano())
	randGen := rand.New(randSource)

	username := make([]byte, length)
	for i := range username {
		username[i] = usernameChars[randGen.Intn(len(usernameChars))]
	}
	return string(username)
}

var illegalChars = regexp.MustCompile(`[<>:"/\\|?*]+`)

func sanitizeFilename(filename string) string {
	filename = illegalChars.ReplaceAllString(filename, "")

	filename = strings.TrimRight(filename, ".")

	if filename == "" {
		filename = "unnamed"
	}

	return filename
}

func isImage(data []byte) bool {
	// PNG文件的前缀字节
	if bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}) {
		return true
	}

	// JPEG文件的前缀字节
	if bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}) {
		return true
	}

	// GIF文件的前缀字节
	if bytes.HasPrefix(data, []byte{0x47, 0x49, 0x46, 0x38}) {
		return true
	}

	// WebP文件的前缀字节
	if bytes.HasPrefix(data, []byte{'R', 'I', 'F', 'F'}) && bytes.HasPrefix(data[8:], []byte{'W', 'E', 'B', 'P'}) {
		return true
	}

	// AVIF文件的前缀字节
	if bytes.HasPrefix(data, []byte{'f', 't', 'y', 'p'}) && bytes.Contains(data[4:], []byte("avif")) {
		return true
	}

	// 其他常见格式可以继续扩展
	return false
}
