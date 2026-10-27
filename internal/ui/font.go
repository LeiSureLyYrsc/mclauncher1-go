package ui

import (
	qt "github.com/mappu/miqt/qt6"
)

// preferredCJKFamilies matches the original 微软雅黑 preference, with Linux fallbacks.
var preferredCJKFamilies = []string{
	"微软雅黑",
	"Microsoft YaHei",
	"Microsoft YaHei UI",
	"Noto Sans CJK SC",
	"Noto Sans CJK",
	"WenQuanYi Micro Hei",
	"WenQuanYi Zen Hei",
	"Source Han Sans SC",
	"Droid Sans Fallback",
	"Sans Serif",
}

// ApplyAppFont picks the first available CJK-capable family.
func ApplyAppFont(_ *qt.QApplication) {
	family := "Noto Sans CJK SC"
	for _, f := range preferredCJKFamilies {
		if qt.QFontDatabase_HasFamily(f) {
			family = f
			break
		}
	}
	font := qt.NewQFont6(family, 10)
	qt.QApplication_SetFont(font)
}

func uiFont(pointSize int, bold bool) *qt.QFont {
	family := "Noto Sans CJK SC"
	for _, f := range preferredCJKFamilies {
		if qt.QFontDatabase_HasFamily(f) {
			family = f
			break
		}
	}
	weight := 400
	if bold {
		weight = 700
	}
	return qt.NewQFont7(family, pointSize, weight)
}
