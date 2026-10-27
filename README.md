# mclauncher1-go

忘却的旋律 **MCLauncher 5.31** 主界面的 Go + **Qt6 QML** 复刻（一比一主窗口 UI）。

> **现阶段目标：UI First（QML）**  
> 下一步：**旧版 Mojang / Yggdrasil（邮箱 + 密码）登录**  
> 再后：Microsoft OAuth 授权流

## 界面约定

- 固定窗口 **720×450**，无边框（`Qt.FramelessWindowHint`），标题 `MClauncher`
- **不要自绘整窗**：背景用 QML `Image`，控件用 `Button` / `ComboBox` / `Text` 等定位叠放
- 中文文案：`启动游戏`、`更新游戏`、`设置`、`打开目录`、页脚 `MineCraft Launcher 5.31 by 忘却的旋律`
- 「设置」为 QML `Dialog`，已预留 **邮箱 / 密码** 与「正版登录测试」按钮（占位，无网络）

## 技术栈

- Go 1.21+（本机测试 Go 1.22+ / 1.27 亦可）
- Qt 6 **QML / Quick Controls**（主界面）+ miqt `qt6/qml`（`QQmlApplicationEngine`）
- Go 仍作进程壳与后续 Mojang 登录后端；Widgets 旧实现保留在 `internal/ui/widgets_legacy/`（不参与默认构建路径）

### Debian / Ubuntu 依赖

```bash
sudo apt install \
  qt6-base-dev qt6-declarative-dev pkg-config build-essential libgl1-mesa-dev \
  qml6-module-qtquick \
  qml6-module-qtquick-controls \
  qml6-module-qtquick-layouts \
  qml6-module-qtquick-window \
  qml6-module-qtquick-dialogs \
  qml6-module-qtqml \
  qml6-module-qtqml-workerscript
```

Go 绑定：[miqt](https://github.com/mappu/miqt)（`github.com/mappu/miqt/qt6` + `…/qt6/qml`）

首次编译 miqt 可能较慢（数分钟）；之后增量构建很快。

## 构建 / 运行

```bash
cd mclauncher1-go
go mod tidy
go build -ldflags '-s -w' -o mclauncher ./cmd/mclauncher
./mclauncher
```

无显示器时可做语法/链接检查（仍需 Qt 库）：

```bash
QT_QPA_PLATFORM=offscreen go build -o mclauncher ./cmd/mclauncher
```

## 目录结构

```
cmd/mclauncher/              入口（加载 QML）
qml/                         Main.qml + SettingsDialog.qml（go:embed）
internal/ui/qmlapp.go        QQmlApplicationEngine 加载器
internal/ui/widgets_legacy/  旧 Widgets 主窗口（参考 / 历史）
assets/                      背景图（data.go base64，构建时解码）
```

## 路线图

1. ✅ 主 UI 720×450 **QML** + 背景图（原 Widgets 方案已迁至 `widgets_legacy/`）
2. ⬜ 旧版 Mojang Yggdrasil 登录（设置里的邮箱/密码）
3. ⬜ 启动 / 更新游戏逻辑
4. ⬜ Microsoft 授权（更后）

## License

UI 复刻仅供学习研究；原启动器与皮肤版权归原作者「忘却的旋律」及权利方。
