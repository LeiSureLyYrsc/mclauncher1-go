import QtQuick
import QtQuick.Controls
import QtQuick.Layouts
import QtQuick.Dialogs

Dialog {
    id: root
    title: "游戏设置"
    modal: true
    anchors.centerIn: Overlay.overlay
    width: 420
    height: 460
    standardButtons: Dialog.NoButton
    closePolicy: Popup.CloseOnEscape | Popup.CloseOnPressOutside

    background: Rectangle {
        color: "#f0f0f0"
        border.color: "#999"
    }

    contentItem: ColumnLayout {
        anchors.fill: parent
        anchors.margins: 12
        spacing: 10

        GroupBox {
            title: "正版登录（旧版 Mojang / Yggdrasil）"
            Layout.fillWidth: true
            ColumnLayout {
                anchors.fill: parent
                spacing: 6
                RowLayout {
                    Label { text: "邮箱"; Layout.preferredWidth: 48 }
                    TextField {
                        id: email
                        Layout.fillWidth: true
                        placeholderText: "帐号邮箱"
                    }
                }
                RowLayout {
                    Label { text: "密码"; Layout.preferredWidth: 48 }
                    TextField {
                        id: pass
                        Layout.fillWidth: true
                        placeholderText: "密码"
                        echoMode: TextInput.Password
                    }
                }
                RowLayout {
                    Button {
                        text: "正版登录测试"
                        onClicked: loginTestDlg.open()
                    }
                    Button {
                        text: "删除正版信息"
                        onClicked: {
                            email.text = ""
                            pass.text = ""
                            clearDlg.open()
                        }
                    }
                    Item { Layout.fillWidth: true }
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
                    Label { text: "java路径"; Layout.preferredWidth: 64 }
                    TextField {
                        id: javaPath
                        Layout.fillWidth: true
                        placeholderText: "java路径 / javaw"
                    }
                    Button {
                        text: "浏览"
                        onClicked: javaDlg.open()
                    }
                }
                RowLayout {
                    Label { text: "最大内存"; Layout.preferredWidth: 64 }
                    SpinBox {
                        id: mem
                        from: 512
                        to: 16384
                        value: 1024
                        editable: true
                    }
                    Label { text: "M" }
                }
                RowLayout {
                    Label { text: "游戏名字"; Layout.preferredWidth: 64 }
                    TextField {
                        id: player
                        Layout.fillWidth: true
                        placeholderText: "输入游戏名字"
                    }
                }
                CheckBox {
                    id: debug
                    text: "调试模式"
                }
            }
        }

        Label {
            Layout.fillWidth: true
            wrapMode: Text.WordWrap
            text: "说明：当前为 UI 壳。下一步实现旧版 Mojang 登录，再后是 Microsoft 授权。"
            color: "#222"
        }

        Item { Layout.fillHeight: true }

        RowLayout {
            Layout.alignment: Qt.AlignRight
            Button {
                text: "保存设置"
                onClicked: {
                    saveDlg.open()
                    root.accept()
                }
            }
            Button {
                text: "取消"
                onClicked: root.reject()
            }
        }
    }

    FileDialog {
        id: javaDlg
        title: "选择java"
        fileMode: FileDialog.OpenFile
        nameFilters: ["Java (java javaw)", "All (*)"]
        onAccepted: {
            var u = selectedFile.toString()
            if (u.indexOf("file://") === 0)
                javaPath.text = decodeURIComponent(u.replace("file://", ""))
            else
                javaPath.text = u
        }
    }

    MessageDialog {
        id: loginTestDlg
        title: "正版登录测试"
        text: "网络登录尚未实现。\n此处预留旧版 Mojang Yggdrasil（邮箱/密码）接口；Microsoft OAuth 将在其后接入。"
        buttons: MessageDialog.Ok
    }
    MessageDialog {
        id: clearDlg
        title: "删除正版信息"
        text: "已清空本地填写（尚未连接存档）。"
        buttons: MessageDialog.Ok
    }
    MessageDialog {
        id: saveDlg
        title: "保存设置"
        text: "设置 UI 已填写，持久化与登录逻辑尚未接入。"
        buttons: MessageDialog.Ok
    }
}
