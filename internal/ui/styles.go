package ui

// Stylesheets approximate the original translucent light controls.
// No full-window paintEvent / QPainter skin — widgets only.

const styleLaunch = `
QPushButton {
	background-color: rgba(245, 245, 245, 230);
	color: #1a1a1a;
	border: 1px solid rgba(180, 180, 180, 200);
	border-radius: 0px;
	font-size: 14px;
	font-weight: 600;
	padding: 0px;
}
QPushButton:hover {
	background-color: rgba(255, 255, 255, 245);
}
QPushButton:pressed {
	background-color: rgba(220, 220, 220, 240);
}
`

const styleSmall = `
QPushButton {
	background-color: rgba(245, 245, 245, 230);
	color: #1a1a1a;
	border: 1px solid rgba(180, 180, 180, 200);
	border-radius: 0px;
	font-size: 11px;
	padding: 0px;
}
QPushButton:hover {
	background-color: rgba(255, 255, 255, 245);
}
QPushButton:pressed {
	background-color: rgba(220, 220, 220, 240);
}
`

const styleCombo = `
QComboBox {
	background-color: rgba(250, 250, 250, 235);
	color: #1a1a1a;
	border: 1px solid rgba(160, 160, 160, 200);
	border-radius: 0px;
	padding-left: 6px;
	font-size: 11px;
}
QComboBox::drop-down {
	border: none;
	width: 18px;
}
QComboBox::down-arrow {
	width: 10px;
	height: 10px;
}
QComboBox QAbstractItemView {
	background: #fafafa;
	selection-background-color: #cce0ff;
	color: #1a1a1a;
}
`

const styleOpenDir = `
QPushButton {
	background: transparent;
	color: #ffffff;
	border: none;
	text-align: left;
	font-size: 11px;
	padding: 0px;
}
QPushButton:hover {
	color: #ffe08a;
	text-decoration: underline;
}
`

const styleFooter = `
QLabel {
	color: rgba(255, 255, 255, 210);
	background: transparent;
	font-size: 10px;
}
`

const styleClose = `
QPushButton {
	background-color: rgba(0, 0, 0, 40);
	color: #ffffff;
	border: 1px solid rgba(255, 255, 255, 80);
	border-radius: 0px;
	font-size: 12px;
	font-weight: bold;
}
QPushButton:hover {
	background-color: #e81123;
	border-color: #e81123;
}
`

const styleSettings = `
QDialog {
	background-color: #f0f0f0;
}
QLabel {
	color: #222;
}
QLineEdit, QComboBox, QSpinBox {
	background: #fff;
	border: 1px solid #aaa;
	padding: 3px;
	min-height: 22px;
}
QPushButton {
	background: #e8e8e8;
	border: 1px solid #999;
	padding: 4px 12px;
	min-height: 24px;
}
QPushButton:hover {
	background: #f5f5f5;
}
QGroupBox {
	font-weight: 600;
	border: 1px solid #bbb;
	margin-top: 8px;
	padding-top: 8px;
}
QGroupBox::title {
	subcontrol-origin: margin;
	left: 8px;
	padding: 0 4px;
}
`
