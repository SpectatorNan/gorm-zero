package queryx

import (
	"strings"
	"unicode"

	"gorm.io/gorm"
)

func WhereOptionalInt64(db *gorm.DB, whereStr string, val *int64) *gorm.DB {
	if val != nil {
		return db.Where(whereStr, val)
	}
	return db
}

func WhereString(db *gorm.DB, whereStr string, fuzz string) *gorm.DB {
	if len(fuzz) > 0 {
		if strings.Contains(whereStr, " like ") {
			return db.Where(whereStr, "%"+fuzz+"%")
		} else {
			return db.Where(whereStr, fuzz)
		}
	}
	return db
}

func WhereOptionalBool(db *gorm.DB, whereStr string, b *bool) *gorm.DB {
	if b != nil {
		return db.Where(whereStr, b)
	}
	return db
}

func WhereUint64(db *gorm.DB, whereStr string, val uint64) *gorm.DB {
	if val != 0 {
		return db.Where(whereStr, val)
	}
	return db
}

func WhereInt64s(db *gorm.DB, whereStr string, val []int64) *gorm.DB {
	if len(val) > 1 {
		return db.Where(whereStr, val)
	} else if len(val) == 1 {
		whereStr = normalizeINToEquals(whereStr)
		return db.Where(whereStr, val[0])
	}
	return db
}

func WhereInt64(db *gorm.DB, whereStr string, val int64) *gorm.DB {
	if val != 0 {
		return db.Where(whereStr, val)
	}
	return db
}

func normalizeINToEquals(s string) string {
	low := strings.ToLower(s)
	for i := len(low) - 2; i >= 0; i-- {
		if low[i:i+2] != "in" {
			continue
		}
		// ensure it's a whole word: previous and next chars are not word chars
		if i > 0 && isWordChar(rune(low[i-1])) {
			continue
		}
		if i+2 < len(low) && isWordChar(rune(low[i+2])) {
			continue
		}
		left := strings.TrimSpace(s[:i])
		right := strings.TrimSpace(s[i+2:])

		if left == "" || right == "" {
			return s
		}
		return left + " = " + right
	}
	return s
}

func isWordChar(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
