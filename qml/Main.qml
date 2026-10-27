import QtQuick
import QtQuick.Controls
import QtQuick.Window
import QtQuick.Dialogs

ApplicationWindow {
    id: win
    title: "MClauncher"
    width: 720
    height: 450
    minimumWidth: 720
    maximumWidth: 720
    minimumHeight: 450
    maximumHeight: 450
    visible: true
    flags: Qt.Window | Qt.FramelessWindowHint
    color: "#00000000"
    font.family: "Noto Sans CJK SC"
    font.pixelSize: 11

    property url bgSource: typeof bgImageUrl !== "undefined" ? bgImageUrl : "bg_main.jpg"
    property string openDirPath: typeof gameDirPath !== "undefined" ? gameDirPath : "."

    // Full-window background via Image (not Canvas paint of the UI).
    Image {
        id: bg
        anchors.fill: parent
        source: win.bgSource
        fillMode: Image.Stretch
        asynchronous: false
        z: 0
    }

    // Drag the frameless window by grabbing empty chrome.
    MouseArea {
        id: dragArea
        anchors.fill: parent
        z: 1
        acceptedButtons: Qt.LeftButton
        property point pressPos: Qt.point(0, 0)
        onPressed: function(mouse) {
            pressPos = Qt.point(mouse.x, mouse.y)
        }
        onPositionChanged: function(mouse) {
            if (pressed) {
                win.x += mouse.x - pressPos.x
                win.y += mouse.y - pressPos.y
            }
        }
    }

    // Classic flat Win32-like button chrome (opaque gray, 1px border, square).
    component WinBtnBg: Rectangle {
        property bool down: false
        property bool hovered: false
        color: down ? "#e1e1e1" : (hovered ? "#f5f5f5" : "#f0f0f0")
        border.color: "#a0a0a0"
        border.width: 1
        radius: 0
    }

    // --- Right control cluster (tightened vs original 720×450 mask) ---
    // Launch widened (≥100) so full 「启动游戏」 fits; side stack flush-top.

    Button {
        id: btnLaunch
        x: 400; y: 280; width: 102; height: 34
        z: 2
        text: "启动游戏"
        font.pixelSize: 13
        font.bold: true
        padding: 0
        leftPadding: 0; rightPadding: 0; topPadding: 0; bottomPadding: 0
        background: WinBtnBg {
            down: btnLaunch.down
            hovered: btnLaunch.hovered
        }
        contentItem: Text {
            text: btnLaunch.text
            font: btnLaunch.font
            color: "#000000"
            horizontalAlignment: Text.AlignHCenter
            verticalAlignment: Text.AlignVCenter
            elide: Text.ElideNone
            clip: false
        }
        onClicked: launchDlg.open()
    }

    Button {
        id: btnUpdate
        x: 506; y: 280; width: 64; height: 16
        z: 2
        text: "更新游戏"
        font.pixelSize: 11
        padding: 0
        leftPadding: 0; rightPadding: 0; topPadding: 0; bottomPadding: 0
        background: WinBtnBg {
            down: btnUpdate.down
            hovered: btnUpdate.hovered
        }
        contentItem: Text {
            text: btnUpdate.text
            font: btnUpdate.font
            color: "#000000"
            horizontalAlignment: Text.AlignHCenter
            verticalAlignment: Text.AlignVCenter
            elide: Text.ElideNone
            clip: false
        }
        onClicked: updateDlg.open()
    }

    Button {
        id: btnSettings
        x: 506; y: 298; width: 64; height: 16
        z: 2
        text: "设置"
        font.pixelSize: 11
        padding: 0
        leftPadding: 0; rightPadding: 0; topPadding: 0; bottomPadding: 0
        background: WinBtnBg {
            down: btnSettings.down
            hovered: btnSettings.hovered
        }
        contentItem: Text {
            text: btnSettings.text
            font: btnSettings.font
            color: "#000000"
            horizontalAlignment: Text.AlignHCenter
            verticalAlignment: Text.AlignVCenter
            elide: Text.ElideNone
            clip: false
        }
        onClicked: settingsDialog.open()
    }

    ComboBox {
        id: combo
        x: 400; y: 318; width: 120; height: 19
        z: 2
        model: ["(默认)", "正式版", "快照预览版"]
        font.pixelSize: 11
        padding: 0
        leftPadding: 0; rightPadding: 0; topPadding: 0; bottomPadding: 0
        background: Rectangle {
            color: "#ffffff"
            border.color: "#a0a0a0"
            border.width: 1
            radius: 0
        }
        contentItem: Text {
            leftPadding: 4
            rightPadding: 18
            text: combo.displayText
            font: combo.font
            color: "#000000"
            verticalAlignment: Text.AlignVCenter
            elide: Text.ElideNone
            clip: true
        }
        indicator: Item {
            width: 16
            height: parent.height
            anchors.right: parent.right
            Text {
                anchors.centerIn: parent
                text: "▼"
                font.pixelSize: 8
                color: "#000000"
            }
        }
        popup: Popup {
            y: combo.height
            width: combo.width
            implicitHeight: Math.min(contentItem.implicitHeight, 120)
            padding: 1
            contentItem: ListView {
                clip: true
                implicitHeight: contentHeight
                model: combo.popup.visible ? combo.delegateModel : null
                currentIndex: combo.highlightedIndex
                ScrollIndicator.vertical: ScrollIndicator {}
            }
            background: Rectangle {
                color: "#ffffff"
                border.color: "#a0a0a0"
                radius: 0
            }
        }
    }

    Text {
        id: openDir
        x: 528; y: 318; width: 72; height: 20
        z: 2
        text: "打开目录"
        color: openDirMa.containsMouse ? "#ffe08a" : "#ffffff"
        font.pixelSize: 11
        font.underline: openDirMa.containsMouse
        verticalAlignment: Text.AlignVCenter
        MouseArea {
            id: openDirMa
            anchors.fill: parent
            cursorShape: Qt.PointingHandCursor
            hoverEnabled: true
            onClicked: {
                var p = win.openDirPath
                if (p.indexOf("file:") === 0)
                    Qt.openUrlExternally(p)
                else
                    Qt.openUrlExternally("file://" + p)
            }
        }
    }

    Text {
        id: footer
        x: 8; y: 428; width: 360; height: 18
        z: 2
        text: "MineCraft Launcher 5.31 by 忘却的旋律"
        color: "#d2ffffff"
        font.pixelSize: 10
        verticalAlignment: Text.AlignVCenter
    }

    Button {
        id: btnClose
        x: 698; y: 4; width: 18; height: 18
        z: 3
        text: "×"
        font.pixelSize: 12
        font.bold: true
        padding: 0
        leftPadding: 0; rightPadding: 0; topPadding: 0; bottomPadding: 0
        background: Rectangle {
            color: btnClose.hovered ? "#e81123" : "#00000000"
            border.width: 0
            radius: 0
        }
        contentItem: Text {
            text: btnClose.text
            font: btnClose.font
            color: "#ffffff"
            horizontalAlignment: Text.AlignHCenter
            verticalAlignment: Text.AlignVCenter
            elide: Text.ElideNone
        }
        onClicked: Qt.quit()
    }

    MessageDialog {
        id: launchDlg
        title: "启动游戏"
        text: "启动逻辑尚未实现。\n下一步将接入旧版 Mojang Yggdrasil（邮箱/密码）登录。"
        buttons: MessageDialog.Ok
    }

    MessageDialog {
        id: updateDlg
        title: "更新游戏"
        text: "更新功能占位（未实现网络下载）。"
        buttons: MessageDialog.Ok
    }

    SettingsDialog {
        id: settingsDialog
    }
}
