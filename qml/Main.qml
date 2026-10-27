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

    // --- Right control cluster (geometry from Widgets 1:1 layout) ---
    // Colors use Qt #AARRGGBB

    Button {
        id: btnLaunch
        x: 400; y: 283; width: 88; height: 30
        z: 2
        text: "启动游戏"
        font.pixelSize: 14
        font.bold: true
        background: Rectangle {
            color: btnLaunch.down ? "#ffdcdcdc" : (btnLaunch.hovered ? "#f5ffffff" : "#e6f5f5f5")
            border.color: "#c8b4b4b4"
            border.width: 1
        }
        contentItem: Text {
            text: btnLaunch.text
            font: btnLaunch.font
            color: "#ff1a1a1a"
            horizontalAlignment: Text.AlignHCenter
            verticalAlignment: Text.AlignVCenter
        }
        onClicked: launchDlg.open()
    }

    Button {
        id: btnUpdate
        x: 492; y: 283; width: 52; height: 14
        z: 2
        text: "更新游戏"
        font.pixelSize: 11
        background: Rectangle {
            color: btnUpdate.down ? "#ffdcdcdc" : (btnUpdate.hovered ? "#f5ffffff" : "#e6f5f5f5")
            border.color: "#c8b4b4b4"
            border.width: 1
        }
        contentItem: Text {
            text: btnUpdate.text
            font: btnUpdate.font
            color: "#ff1a1a1a"
            horizontalAlignment: Text.AlignHCenter
            verticalAlignment: Text.AlignVCenter
        }
        onClicked: updateDlg.open()
    }

    Button {
        id: btnSettings
        x: 492; y: 299; width: 52; height: 14
        z: 2
        text: "设置"
        font.pixelSize: 11
        background: Rectangle {
            color: btnSettings.down ? "#ffdcdcdc" : (btnSettings.hovered ? "#f5ffffff" : "#e6f5f5f5")
            border.color: "#c8b4b4b4"
            border.width: 1
        }
        contentItem: Text {
            text: btnSettings.text
            font: btnSettings.font
            color: "#ff1a1a1a"
            horizontalAlignment: Text.AlignHCenter
            verticalAlignment: Text.AlignVCenter
        }
        onClicked: settingsDialog.open()
    }

    ComboBox {
        id: combo
        x: 401; y: 320; width: 100; height: 16
        z: 2
        model: ["(默认)", "正式版", "快照预览版"]
        font.pixelSize: 11
        background: Rectangle {
            color: "#ebfafafa"
            border.color: "#c8a0a0a0"
            border.width: 1
        }
        contentItem: Text {
            leftPadding: 6
            text: combo.displayText
            font: combo.font
            color: "#ff1a1a1a"
            verticalAlignment: Text.AlignVCenter
            elide: Text.ElideRight
        }
        indicator: Item {
            width: 18
            height: parent.height
            anchors.right: parent.right
            Text {
                anchors.centerIn: parent
                text: "▾"
                font.pixelSize: 10
                color: "#ff333333"
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
                color: "#fffafafa"
                border.color: "#ffaaaaaa"
            }
        }
    }

    Text {
        id: openDir
        x: 510; y: 318; width: 72; height: 18
        z: 2
        text: "打开目录"
        color: openDirMa.containsMouse ? "#ffffe08a" : "#ffffffff"
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
        x: 692; y: 4; width: 22; height: 20
        z: 3
        text: "×"
        font.pixelSize: 12
        font.bold: true
        background: Rectangle {
            color: btnClose.hovered ? "#ffe81123" : "#28000000"
            border.color: btnClose.hovered ? "#ffe81123" : "#50ffffff"
            border.width: 1
        }
        contentItem: Text {
            text: btnClose.text
            font: btnClose.font
            color: "#ffffffff"
            horizontalAlignment: Text.AlignHCenter
            verticalAlignment: Text.AlignVCenter
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
