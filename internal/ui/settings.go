package ui

import (
	qt "github.com/mappu/miqt/qt6"
)

// OpenSettingsDialog shows the settings shell.
// Includes 邮箱/密码 fields reserved for legacy Mojang Yggdrasil login (not wired yet).
func OpenSettingsDialog(parent *qt.QWidget) {
	dlg := qt.NewQDialog(parent)
	dlg.SetWindowTitle("游戏设置")
	dlg.SetModal(true)
	dlg.SetFixedSize2(420, 460)
	dlg.SetStyleSheet(styleSettings)

	root := qt.NewQVBoxLayout(dlg.QWidget)
	root.SetContentsMargins(12, 12, 12, 12)

	// --- 正版登录 (Mojang Yggdrasil placeholder) ---
	loginBox := qt.NewQGroupBox4("正版登录（旧版 Mojang / Yggdrasil）", dlg.QWidget)
	loginForm := qt.NewQFormLayout(loginBox.QWidget)

	email := qt.NewQLineEdit(loginBox.QWidget)
	email.SetPlaceholderText("帐号邮箱")
	loginForm.AddRow3("邮箱", email.QWidget)

	pass := qt.NewQLineEdit(loginBox.QWidget)
	pass.SetPlaceholderText("密码")
	pass.SetEchoMode(qt.QLineEdit__Password)
	loginForm.AddRow3("密码", pass.QWidget)

	loginBtns := qt.NewQHBoxLayout2()
	btnTestLogin := qt.NewQPushButton5("正版登录测试", loginBox.QWidget)
	btnClearLogin := qt.NewQPushButton5("删除正版信息", loginBox.QWidget)
	loginBtns.AddWidget(btnTestLogin.QWidget)
	loginBtns.AddWidget(btnClearLogin.QWidget)
	loginBtns.AddStretch()
	loginForm.AddRow4("", loginBtns.QLayout)

	btnTestLogin.OnClicked(func() {
		qt.QMessageBox_Information(dlg.QWidget, "正版登录测试",
			"网络登录尚未实现。\n此处预留旧版 Mojang Yggdrasil（邮箱/密码）接口；Microsoft OAuth 将在其后接入。")
	})
	btnClearLogin.OnClicked(func() {
		email.SetText("")
		pass.SetText("")
		qt.QMessageBox_Information(dlg.QWidget, "删除正版信息", "已清空本地填写（尚未连接存档）。")
	})

	root.AddWidget(loginBox.QWidget)

	// --- 游戏设置 ---
	gameBox := qt.NewQGroupBox4("游戏设置", dlg.QWidget)
	gameForm := qt.NewQFormLayout(gameBox.QWidget)

	javaPath := qt.NewQLineEdit(gameBox.QWidget)
	javaPath.SetPlaceholderText("java路径 / javaw")
	browse := qt.NewQPushButton5("浏览", gameBox.QWidget)
	javaRow := qt.NewQHBoxLayout2()
	javaRow.AddWidget2(javaPath.QWidget, 1)
	javaRow.AddWidget(browse.QWidget)
	gameForm.AddRow4("java路径", javaRow.QLayout)
	browse.OnClicked(func() {
		path := qt.QFileDialog_GetOpenFileName4(dlg.QWidget, "选择java", "", "Java (java javaw);;All (*)")
		if path != "" {
			javaPath.SetText(path)
		}
	})

	mem := qt.NewQSpinBox(gameBox.QWidget)
	mem.SetRange(512, 16384)
	mem.SetValue(1024)
	mem.SetSuffix(" M")
	gameForm.AddRow3("最大内存", mem.QWidget)

	player := qt.NewQLineEdit(gameBox.QWidget)
	player.SetPlaceholderText("输入游戏名字")
	gameForm.AddRow3("游戏名字", player.QWidget)

	debug := qt.NewQCheckBox4("调试模式", gameBox.QWidget)
	gameForm.AddRowWithWidget(debug.QWidget)

	root.AddWidget(gameBox.QWidget)

	hint := qt.NewQLabel5("说明：当前为 UI 壳。下一步实现旧版 Mojang 登录，再后是 Microsoft 授权。", dlg.QWidget)
	hint.SetWordWrap(true)
	root.AddWidget(hint.QWidget)

	root.AddStretch()

	buttons := qt.NewQDialogButtonBox7(qt.QDialogButtonBox__Save|qt.QDialogButtonBox__Cancel, dlg.QWidget)
	// Relabel Save → 保存设置
	if saveBtn := buttons.Button(qt.QDialogButtonBox__Save); saveBtn != nil {
		saveBtn.SetText("保存设置")
	}
	if cancelBtn := buttons.Button(qt.QDialogButtonBox__Cancel); cancelBtn != nil {
		cancelBtn.SetText("取消")
	}
	buttons.OnAccepted(func() {
		qt.QMessageBox_Information(dlg.QWidget, "保存设置", "设置 UI 已填写，持久化与登录逻辑尚未接入。")
		dlg.Accept()
	})
	buttons.OnRejected(func() { dlg.Reject() })
	root.AddWidget(buttons.QWidget)

	dlg.Exec()
}
