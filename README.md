# mclauncher1-go

[![Build](https://github.com/LeiSureLyYrsc/mclauncher1-go/actions/workflows/build.yml/badge.svg)](https://github.com/LeiSureLyYrsc/mclauncher1-go/actions/workflows/build.yml)


忘却的旋律 **MCLauncher 5.31** 主界面的 Go + **Qt6 QML** 复刻，并实现核心启动能力。

## 功能

- **下载源**：BMCLAPI（`https://bmclapi2.bangbang93.com`）——版本清单、客户端、依赖库、资源文件、authlib-injector 镜像
- **登录**：
  - **离线登录**（默认，无需配置；默认名 `Player`，离线 UUID）
  - **Authlib Injector** 外置登录（Yggdrasil 兼容：`/authserver/authenticate` + `/authserver/refresh`）。**外置登录只保存 token，不保存密码**（方案 1：成功后仅持久化 accessToken / clientToken / 角色与 API 地址、显示用用户名）
  - **微软登录**：TODO（尚未实现）
- **版本**：从 BMCLAPI 拉取完整版本列表；支持正式版 / 快照 / 全部筛选；「更新游戏」安装所选版本
- **Java**：自动嗅探 `JAVA_HOME`、PATH 与常见安装目录
- **启动**：组装 classpath / natives / JVM·游戏参数；Authlib 模式自动附加 `-javaagent:authlib-injector.jar=API_URL`

## 界面约定

- 固定窗口 **720×450**，无边框，标题 `MClauncher`
- **不要自绘整窗**：背景用 QML `Image`，控件定位叠放
- 中文 UI（简体）

## 技术栈

- Go 1.21+ / Qt 6 QML + [miqt](https://github.com/mappu/miqt)
- 配置保存在 `.minecraft/.mclauncher/config.json`

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

## 构建 / 运行

```bash
cd mclauncher1-go
go mod tidy
go build -ldflags '-s -w' -o mclauncher ./cmd/mclauncher
./mclauncher
```

## 目录结构

```
cmd/mclauncher/           入口
qml/                      Main.qml + SettingsDialog.qml
internal/ui/              QML 加载与 PropertyMap 桥接
internal/bmclapi/         BMCLAPI 清单 / 下载 / 安装
internal/auth/            离线 + Authlib Injector
internal/java/            Java 嗅探
internal/launch/          参数组装与进程启动
internal/config/          本地配置
assets/                   背景图
```

## 使用提示

1. 首次打开会从 BMCLAPI 加载版本列表；在下拉框选择版本
2. 点击「更新游戏」下载该版本（客户端、库、资源）
3. 未配置 Authlib 时直接「启动游戏」即为离线模式
4. 设置中可切换 Authlib、填写 API 地址并「登录测试」（密码仅用于当次认证，不写入 config.json）；下次启动若令牌有效则刷新使用，刷新失败需重新登录；可「嗅探」Java


## CI / 自动构建

GitHub Actions（[`.github/workflows/build.yml`](.github/workflows/build.yml)）在以下事件自动构建：

- **push**：`main` 及任意分支
- **pull_request**
- **tags**：`v*`（例如 `v0.1.0`）

在 `ubuntu-latest` 上安装 Qt6 + miqt QML 依赖，按 `go.mod` 安装 Go，执行：

```bash
go build -ldflags '-s -w' -o mclauncher ./cmd/mclauncher
```

产物上传为 Artifact（`mclauncher-linux-amd64`）。打 `v*` 标签时会将 Linux 二进制挂到对应 Release。Windows 交叉编译需 Qt 工具链，暂未纳入 CI。

## License

UI 复刻仅供学习研究；原启动器与皮肤版权归原作者「忘却的旋律」及权利方。
