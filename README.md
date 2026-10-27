# mclauncher1-go

忘却的旋律 **MCLauncher 5.31** 主界面的 Go + Qt6 复刻（一比一主窗口 UI）。

> **现阶段目标：UI First**  
> 下一步：**旧版 Mojang / Yggdrasil（邮箱 + 密码）登录**  
> 再后：Microsoft OAuth 授权流

## 界面约定

- 固定窗口 **720×450**，无边框（`FramelessWindowHint`），标题 `MClauncher`
- **不要自绘整窗**：背景用 `QLabel.SetPixmap`，控件用 `QPushButton` / `QComboBox` / `QLabel` + stylesheet
- 中文文案：`启动游戏`、`更新游戏`、`设置`、`打开目录`、页脚 `MineCraft Launcher 5.31 by 忘却的旋律`
- 「设置」对话框已预留 **邮箱 / 密码** 与「正版登录测试」按钮（占位，无网络）

## 依赖

- Go 1.21+（本机测试 Go 1.22+ / 1.27 亦可）
- Qt 6 Widgets 开发包（Debian/Ubuntu）：

```bash
sudo apt install qt6-base-dev pkg-config build-essential libgl1-mesa-dev
```

- Go 绑定：[miqt](https://github.com/mappu/miqt)（`github.com/mappu/miqt/qt6`）

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
cmd/mclauncher/     入口
internal/ui/        主窗口 + 设置对话框（widgets + QSS）
assets/             背景图（data.go base64，构建时解码）
```

## 路线图

1. ✅ 主 UI 720×450 widgets + 背景图
2. ⬜ 旧版 Mojang Yggdrasil 登录（设置里的邮箱/密码）
3. ⬜ 启动 / 更新游戏逻辑
4. ⬜ Microsoft 授权（更后）

## License

UI 复刻仅供学习研究；原启动器与皮肤版权归原作者「忘却的旋律」及权利方。
