# dockerCopilot
<a href="https://www.gnu.org/licenses/agpl-3.0.en.html">
    <img alt="License: AGPLv3" src="https://shields.io/badge/License-AGPL%20v3-blue.svg">
  </a>

# 介绍

一个主打便捷的docker容器管理工具，现在已经支持所有平台。
已经实现：
1. 一键更新容器
2. 指定镜像和tag更新
3. 启动、停止、重启容器
4. 重命名容器
5. 删除无TAG镜像
6. 删除未使用镜像
7. 更新进度查看
8. 备份容器设置
9. 恢复容器设置

## 使用

docker compose 安装

```
services:
  dockercopilot:
    container_name: dockercopilot
    restart: always
    privileged: true
    network_mode: bridge
    ports:
      - 12712:12712
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./data:/data
      - ./dockerCopilot.yaml:/app/etc/dockerCopilot.yaml
    environment:
      - TZ=Asia/Shanghai
      - DOCKER_HOST=unix:///var/run/docker.sock
      - NO_PROXY=localhost,127.0.0.1,172.17.0.0/16,192.168.0.0/16
      - secretKey=密码，不少于八位且非纯数字
    image: 0nlylty/dockercopilot:latest
```

### 代理配置

如果你的服务器无法直接访问 Docker Hub，可以通过配置代理来解决。

**方式一：配置文件（推荐）**

挂载 `dockerCopilot.yaml`，取消 Proxy 部分注释并修改：

```yaml
Proxy:
  Enable: true
  Http: "http://172.17.0.1:7890"    # 代理地址，172.17.0.1 为 Docker 网关 IP
  Https: "http://172.17.0.1:7890"
  NoProxy: "localhost,127.0.0.1,172.17.0.0/16,192.168.0.0/16"
  InsecureSkipVerify: false
```

> 代理地址说明：
> - `host.docker.internal` — Docker Desktop 可用
> - `172.17.0.1` — Linux Docker 默认网桥网关（推荐）
> - 宿主机局域网 IP — 通用方式
>
> 配置代理后，镜像更新检查会直接通过代理访问 `index.docker.io`，自动跳过国内镜像源遍历。

**方式二：环境变量**

```yaml
environment:
  - HTTP_PROXY=http://172.17.0.1:7890
  - HTTPS_PROXY=http://172.17.0.1:7890
  - NO_PROXY=localhost,127.0.0.1,172.17.0.0/16,192.168.0.0/16
```

> 环境变量方式影响所有网络请求，包括 Docker SDK 连接。建议优先使用配置文件方式。

## 开发环境

go版本：1.21+

