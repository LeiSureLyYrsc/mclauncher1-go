import QtQuick
import QtQuick.Controls
import QtQuick.Layouts
import QtQuick.Dialogs

Dialog {
    id: root
    title: "游戏设置"
    modal: true
    anchors.centerIn: Overlay.overlay
    width: 440
    height: 520
    standardButtons: Dialog.NoButton
    closePolicy: Popup.CloseOnEscape | Popup.CloseOnPressOutside

    property var L: typeof launcher !== "undefined" ? launcher : null

    background: Rectangle {
        color: "#f0f0f0"
        border.color: "#999"
    }

    onAboutToShow: {
        if (!L) return
        if (L.authMode === "authlib")
            modeAuthlib.checked = true
        else
            modeOffline.checked = true
        offlineName.text = L.offlineName || "Player"
        authlibAPI.text = L.authlibAPI || ""
        authlibUser.text = L.authlibUser || ""
        // Password is ephemeral — never reload from config (scheme 1).
        authlibPass.text = ""
        javaPath.text = L.javaPath || ""
        mem.value = L.maxMemoryMB > 0 ? L.maxMemoryMB : 2048
        debug.checked = (L.debug === true || L.debug === "true" || L.debug === "1")
        if (L.versionFilter === "all")
            filterAll.checked = true
        else if (L.versionFilter === "snapshot")
            filterSnap.checked = true
        else
            filterRelease.checked = true
    }

    contentItem: ColumnLayout {
        anchors.fill: parent
        anchors.margins: 12
        spacing: 8

        GroupBox {
            title: "账户"
            Layout.fillWidth: true
            ColumnLayout {
                anchors.fill: parent
                spacing: 6
                RowLayout {
                    RadioButton {
                        id: modeOffline
                        text: "离线登录"
                        checked: true
                    }
                    RadioButton {
                        id: modeAuthlib
                        text: "Authlib Injector"
                    }
                }
                RowLayout {
                    visible: modeOffline.checked
                    Label { text: "游戏名"; Layout.preferredWidth: 72 }
                    TextField {
                        id: offlineName
                        Layout.fillWidth: true
                        placeholderText: "Player（默认离线名）"
                    }
                }
                ColumnLayout {
                    visible: modeAuthlib.checked
                    spacing: 6
                    RowLayout {
                        Label { text: "API 地址"; Layout.preferredWidth: 72 }
                        TextField {
                            id: authlibAPI
                            Layout.fillWidth: true
                            placeholderText: "https://example.com/api/yggdrasil"
                        }
                    }
                    RowLayout {
                        Label { text: "用户名"; Layout.preferredWidth: 72 }
                        TextField {
                            id: authlibUser
                            Layout.fillWidth: true
                            placeholderText: "邮箱 / 用户名"
                        }
                    }
                    RowLayout {
                        Label { text: "密码"; Layout.preferredWidth: 72 }
                        TextField {
                            id: authlibPass
                            Layout.fillWidth: true
                            placeholderText: "仅用于本次登录，不会保存"
                            echoMode: TextInput.Password
                        }
                    }
                    Label {
                        Layout.fillWidth: true
                        wrapMode: Text.WordWrap
                        font.pixelSize: 11
                        color: "#555"
                        text: {
                            var pn = L && L.profileName ? L.profileName : ""
                            if (pn)
                                return "已登录角色：" + pn + "（仅保存令牌，不保存密码）"
                            return "外置登录只保存 access/client 令牌与角色信息，不保存密码。"
                        }
                    }
                    RowLayout {
                        Button {
                            text: "登录测试"
                            onClicked: {
                                pushToLauncher()
                                if (L) L.cmd = "testLogin"
                                // Keep field empty after submit so reopen/clear never shows a saved password.
                                authlibPass.text = ""
                            }
                        }
                        Button {
                            text: "清除登录"
                            onClicked: {
                                if (L) L.cmd = "clearAuth"
                                authlibUser.text = ""
                                authlibPass.text = ""
                                modeOffline.checked = true
                            }
                        }
                        Item { Layout.fillWidth: true }
                    }
                }
            }
        }

        GroupBox {
            title: "游戏设置"
            Layout.fillWidth: true
            ColumnLayout {
                anchors.fill: parent
                spacing: 6
                RowLayout {
                    Label { text: "Java 路径"; Layout.preferredWidth: 72 }
                    TextField {
                        id: javaPath
                        Layout.fillWidth: true
                        placeholderText: "自动嗅探 / 手动指定"
                    }
                    Button {
                        text: "嗅探"
                        onClicked: {
                            if (L) L.cmd = "sniffJava"
                            Qt.callLater(function() {
                                if (L) javaPath.text = L.javaPath || javaPath.text
                            })
                        }
                    }
                    Button {
                        text: "浏览"
                        onClicked: javaDlg.open()
                    }
                }
                RowLayout {
                    Label { text: "最大内存"; Layout.preferredWidth: 72 }
                    SpinBox {
                        id: mem
                        from: 512
                        to: 16384
                        value: 2048
                        editable: true
                        stepSize: 256
                    }
                    Label { text: "M" }
                }
                RowLayout {
                    Label { text: "版本筛选"; Layout.preferredWidth: 72 }
                    RadioButton { id: filterRelease; text: "正式版"; checked: true }
                    RadioButton { id: filterSnap; text: "快照" }
                    RadioButton { id: filterAll; text: "全部" }
                }
                CheckBox {
                    id: debug
                    text: "调试模式（输出到控制台）"
                }
            }
        }

        Label {
            Layout.fillWidth: true
            wrapMode: Text.WordWrap
            text: "下载源：BMCLAPI（bmclapi2）。未配置 Authlib 时默认离线登录。外置登录只保存 token，不保存密码。微软登录稍后实现。"
            color: "#222"
        }

        Label {
            Layout.fillWidth: true
            wrapMode: Text.WordWrap
            text: L ? ("状态：" + (L.status || "")) : ""
            color: "#444"
            font.pixelSize: 11
        }

        Item { Layout.fillHeight: true }

        RowLayout {
            Layout.alignment: Qt.AlignRight
            Button {
                text: "保存设置"
                onClicked: {
                    pushToLauncher()
                    if (L) L.cmd = "saveSettings"
                    root.accept()
                }
            }
            Button {
                text: "取消"
                onClicked: root.reject()
            }
        }
    }

    function pushToLauncher() {
        if (!L) return
        L.authMode = modeAuthlib.checked ? "authlib" : "offline"
        L.offlineName = offlineName.text || "Player"
        L.authlibAPI = authlibAPI.text
        L.authlibUser = authlibUser.text
        // Ephemeral: password only for authenticate; never persisted to config.json.
        L.authlibPass = authlibPass.text
        L.javaPath = javaPath.text
        L.maxMemoryMB = mem.value
        L.debug = debug.checked
        var f = "release"
        if (filterSnap.checked) f = "snapshot"
        if (filterAll.checked) f = "all"
        L.versionFilter = f
    }

    FileDialog {
        id: javaDlg
        title: "选择 Java"
        fileMode: FileDialog.OpenFile
        nameFilters: ["Java (java javaw java.exe javaw.exe)", "All (*)"]
        onAccepted: {
            var u = selectedFile.toString()
            if (u.indexOf("file://") === 0)
                javaPath.text = decodeURIComponent(u.replace(/file:\/\//, ""))
            else
                javaPath.text = u
        }
    }
}
