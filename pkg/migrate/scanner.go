package migrate

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// versionFilePattern 匹配迁移文件名格式：{version}_{desc}.up.sql 或 .down.sql
// 例如：20260420_001_init_schema.up.sql
var versionFilePattern = regexp.MustCompile(
	`^(\d{8}_\d{3})_([a-zA-Z0-9_]+)\.(up|down)\.sql$`,
)

// Scan 扫描 dir 目录，解析所有符合命名规范的迁移文件。
// 返回的 Migration 列表按 Version 字典序升序排列。
func Scan(dir string) ([]*Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read migrations dir %q: %w", dir, err)
	}

	index := make(map[string]*Migration)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		m := versionFilePattern.FindStringSubmatch(name)
		if m == nil {
			continue // 忽略不符合规范的文件
		}

		version := m[1]
		desc := m[2]
		direction := m[3] // "up" or "down"

		mig, ok := index[version]
		if !ok {
			mig = &Migration{
				Version:     version,
				Description: desc,
			}
			index[version] = mig
		}

		absPath := filepath.Join(dir, name)
		switch direction {
		case "up":
			mig.UpFile = absPath
		case "down":
			mig.DownFile = absPath
		}
	}

	// 计算 up 文件校验和
	migrations := make([]*Migration, 0, len(index))
	for _, mig := range index {
		if mig.UpFile == "" {
			continue // 无 up 文件的版本跳过
		}
		checksum, err := fileChecksum(mig.UpFile)
		if err != nil {
			return nil, fmt.Errorf("compute checksum for %q: %w", mig.UpFile, err)
		}
		mig.Checksum = checksum
		migrations = append(migrations, mig)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// fileChecksum 计算文件内容的 SHA256 十六进制字符串。
func fileChecksum(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum), nil
}

// nextVersion 根据已有文件计算下一个版本号（格式 YYYYMMDD_NNN）。
func nextVersion(dir string, date string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("read migrations dir: %w", err)
	}

	// 收集当天已有序号
	prefix := date + "_"
	maxSeq := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		var seq int
		rest := strings.TrimPrefix(name, prefix)
		_, _ = fmt.Sscanf(rest, "%d", &seq)
		if seq > maxSeq {
			maxSeq = seq
		}
	}

	return fmt.Sprintf("%s_%03d", date, maxSeq+1), nil
}
